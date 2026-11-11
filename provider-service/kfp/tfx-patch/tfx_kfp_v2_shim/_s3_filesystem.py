"""boto3-backed TFX fileio plugin for ``s3://`` paths (patch 10).

S3 has no directories: a "directory" exists when any object has its path as a
prefix, ``makedirs`` is a no-op and ``rmtree`` deletes every object under the
prefix.
"""

from __future__ import annotations

import fnmatch
import functools
import io
from types import SimpleNamespace

SCHEME = "s3://"
# TFX's TensorflowFilesystem also claims s3:// at priority 10; lower wins.
PRIORITY = 5
_DELETE_BATCH = 1000
_GLOB_CHARS = "*?["


@functools.lru_cache(maxsize=1)
def _client():
    import boto3
    return boto3.client("s3")


def _split(path) -> tuple[str, str]:
    if isinstance(path, bytes):
        path = path.decode("utf-8")
    if not path.startswith(SCHEME):
        raise ValueError(f"Not an S3 path: {path!r}")
    bucket, _, key = path[len(SCHEME):].partition("/")
    return bucket, key


def _dir_prefix(key: str) -> str:
    return key.rstrip("/") + "/" if key else ""


def _is_not_found(exc: Exception) -> bool:
    code = getattr(exc, "response", {}).get("Error", {}).get("Code")
    return code in ("NoSuchKey", "NotFound", "404")


def _pages(bucket: str, prefix: str, delimiter: str | None = None):
    kwargs = {"Bucket": bucket, "Prefix": prefix}
    if delimiter:
        kwargs["Delimiter"] = delimiter
    while True:
        page = _client().list_objects_v2(**kwargs)
        yield page
        if not page.get("IsTruncated"):
            return
        kwargs["ContinuationToken"] = page["NextContinuationToken"]


def _keys(bucket: str, prefix: str):
    for page in _pages(bucket, prefix):
        for obj in page.get("Contents", []):
            yield obj["Key"]


def _children(bucket: str, key: str) -> tuple[list[str], list[str]]:
    """Return the immediate sub-directory and file names under *key*."""
    prefix = _dir_prefix(key)
    dirs, files = [], []
    for page in _pages(bucket, prefix, delimiter="/"):
        for common in page.get("CommonPrefixes", []):
            dirs.append(common["Prefix"][len(prefix):].rstrip("/"))
        for obj in page.get("Contents", []):
            name = obj["Key"][len(prefix):]
            if name:
                files.append(name)
    return dirs, files


def _head(bucket: str, key: str):
    """Return head_object's response, or None if the object does not exist."""
    if not key:
        return None
    try:
        return _client().head_object(Bucket=bucket, Key=key)
    except Exception as exc:
        if _is_not_found(exc):
            return None
        raise


class _Writer(io.BytesIO):
    """Buffers writes and uploads the object on close."""

    def __init__(self, bucket: str, key: str):
        super().__init__()
        self._bucket = bucket
        self._key = key

    def close(self):
        if not self.closed:
            _client().put_object(Bucket=self._bucket, Key=self._key, Body=self.getvalue())
        super().close()


def build_s3_filesystem(base_cls: type, not_found_error: type[Exception]) -> type:
    """Build the plugin class on TFX's ``Filesystem`` and ``NotFoundError``."""

    class S3Filesystem(base_cls):
        SUPPORTED_SCHEMES = [SCHEME]

        @staticmethod
        def open(name, mode="r"):
            bucket, key = _split(name)
            if "a" in mode or "+" in mode:
                raise ValueError(f"Mode {mode!r} is not supported for {name!r}")
            if "w" in mode:
                writer = _Writer(bucket, key)
                return writer if "b" in mode else io.TextIOWrapper(writer, encoding="utf-8")
            try:
                body = _client().get_object(Bucket=bucket, Key=key)["Body"].read()
            except Exception as exc:
                if _is_not_found(exc):
                    raise not_found_error(f"{name} not found") from exc
                raise
            return io.BytesIO(body) if "b" in mode else io.StringIO(body.decode("utf-8"))

        @staticmethod
        def copy(src, dst, overwrite=False):
            if not overwrite and S3Filesystem.exists(dst):
                raise OSError(f"Destination file {dst!r} already exists")
            src_bucket, src_key = _split(src)
            dst_bucket, dst_key = _split(dst)
            if _head(src_bucket, src_key) is None:
                raise not_found_error(f"{src} not found")
            _client().copy_object(
                CopySource={"Bucket": src_bucket, "Key": src_key},
                Bucket=dst_bucket,
                Key=dst_key,
            )

        @staticmethod
        def exists(path):
            bucket, key = _split(path)
            return _head(bucket, key) is not None or S3Filesystem.isdir(path)

        @staticmethod
        def glob(pattern):
            bucket, key = _split(pattern)
            literal = key
            for char in _GLOB_CHARS:
                literal = literal.split(char, 1)[0]
            return [
                f"{SCHEME}{bucket}/{k}"
                for k in _keys(bucket, literal)
                if fnmatch.fnmatchcase(k, key)
            ]

        @staticmethod
        def isdir(path):
            bucket, key = _split(path)
            if not key:
                return True
            page = _client().list_objects_v2(Bucket=bucket, Prefix=_dir_prefix(key), MaxKeys=1)
            return bool(page.get("Contents") or page.get("CommonPrefixes"))

        @staticmethod
        def listdir(path):
            bucket, key = _split(path)
            dirs, files = _children(bucket, key)
            if not dirs and not files and not S3Filesystem.isdir(path):
                raise not_found_error(f"{path} not found")
            return dirs + files

        @staticmethod
        def makedirs(path):
            _split(path)

        @staticmethod
        def mkdir(path):
            _split(path)

        @staticmethod
        def remove(path):
            bucket, key = _split(path)
            if _head(bucket, key) is None:
                raise not_found_error(f"{path} not found")
            _client().delete_object(Bucket=bucket, Key=key)

        @staticmethod
        def rename(src, dst, overwrite=False):
            S3Filesystem.copy(src, dst, overwrite=overwrite)
            S3Filesystem.remove(src)

        @staticmethod
        def rmtree(path):
            bucket, key = _split(path)
            keys = list(_keys(bucket, _dir_prefix(key)))
            if not keys:
                raise not_found_error(f"{path} not found")
            for i in range(0, len(keys), _DELETE_BATCH):
                _client().delete_objects(
                    Bucket=bucket,
                    Delete={"Objects": [{"Key": k} for k in keys[i:i + _DELETE_BATCH]]},
                )

        @staticmethod
        def stat(path):
            bucket, key = _split(path)
            head = _head(bucket, key)
            if head is not None:
                return SimpleNamespace(
                    length=head["ContentLength"],
                    mtime_nsec=int(head["LastModified"].timestamp() * 1e9),
                    is_directory=False,
                )
            if S3Filesystem.isdir(path):
                return SimpleNamespace(length=0, mtime_nsec=0, is_directory=True)
            raise not_found_error(f"{path} not found")

        @staticmethod
        def walk(top, topdown=True, onerror=None):
            bucket, key = _split(top)
            top = f"{SCHEME}{bucket}/{key}".rstrip("/")
            dirs, files = _children(bucket, key)
            if key and not dirs and not files:
                return
            if topdown:
                yield top, dirs, files
            for name in dirs:
                yield from S3Filesystem.walk(f"{top}/{name}", topdown, onerror)
            if not topdown:
                yield top, dirs, files

    return S3Filesystem
