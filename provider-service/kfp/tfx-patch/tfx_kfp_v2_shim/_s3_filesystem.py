"""boto3-backed TFX fileio plugin for ``s3://`` paths (patch 10).

S3 has no directories: a "directory" exists when any object has its path as a
prefix, ``makedirs`` is a no-op and ``rmtree`` deletes every object under the
prefix.
"""

from __future__ import annotations

import fnmatch
import functools
import io
from collections.abc import Callable, Iterator
from dataclasses import dataclass
from typing import IO, Any, Union

PathType = Union[str, bytes]

SCHEME = "s3://"
# TFX's TensorflowFilesystem also claims s3:// at priority 10; lower wins.
PRIORITY = 5
_DELETE_BATCH_SIZE = 1000
_GLOB_CHARS = "*?["
_NOT_FOUND_CODES = frozenset({"NoSuchKey", "NotFound", "404"})


@dataclass(frozen=True)
class S3Stat:
    """The subset of ``tf.io.gfile.stat``'s result that TFX reads."""

    length: int
    mtime_nsec: int
    is_directory: bool


@functools.lru_cache(maxsize=1)
def _client() -> Any:
    import boto3

    return boto3.client("s3")


@functools.lru_cache(maxsize=1)
def _client_error() -> type[Exception]:
    from botocore.exceptions import ClientError

    return ClientError


def _split(path: PathType) -> tuple[str, str]:
    if isinstance(path, bytes):
        path = path.decode("utf-8")
    if not path.startswith(SCHEME):
        raise ValueError(f"Not an S3 path: {path!r}")
    bucket, _, key = path.removeprefix(SCHEME).partition("/")
    return bucket, key


def _dir_prefix(key: str) -> str:
    return key.rstrip("/") + "/" if key else ""


def _is_not_found(exc: Exception) -> bool:
    code = getattr(exc, "response", {}).get("Error", {}).get("Code")
    return code in _NOT_FOUND_CODES


def _pages(bucket: str, prefix: str, delimiter: str | None = None) -> Iterator[dict]:
    kwargs: dict[str, Any] = {"Bucket": bucket, "Prefix": prefix}
    if delimiter:
        kwargs["Delimiter"] = delimiter
    while True:
        page = _client().list_objects_v2(**kwargs)
        yield page
        if not page.get("IsTruncated"):
            return
        kwargs["ContinuationToken"] = page["NextContinuationToken"]


def _keys(bucket: str, prefix: str) -> Iterator[str]:
    for page in _pages(bucket, prefix):
        for obj in page.get("Contents", []):
            yield obj["Key"]


def _children(bucket: str, key: str) -> tuple[list[str], list[str]]:
    """Return the immediate sub-directory and file names under *key*."""
    prefix = _dir_prefix(key)
    dirs: list[str] = []
    files: list[str] = []
    for page in _pages(bucket, prefix, delimiter="/"):
        for common in page.get("CommonPrefixes", []):
            dirs.append(common["Prefix"].removeprefix(prefix).rstrip("/"))
        for obj in page.get("Contents", []):
            name = obj["Key"].removeprefix(prefix)
            if name:
                files.append(name)
    return dirs, files


def _head(bucket: str, key: str) -> dict | None:
    """Return head_object's response, or None if the object does not exist."""
    if not key:
        return None
    try:
        return _client().head_object(Bucket=bucket, Key=key)
    except _client_error() as exc:
        if _is_not_found(exc):
            return None
        raise


class _Writer(io.BytesIO):
    """Buffers writes and uploads the object on close.

    Leaving a ``with`` block because of an exception discards the buffer, so a
    failed write never replaces the object with partial content.
    """

    def __init__(self, bucket: str, key: str) -> None:
        super().__init__()
        self._bucket = bucket
        self._key = key
        self._discard = False

    def discard(self) -> None:
        self._discard = True

    def __exit__(self, exc_type, exc, tb) -> None:
        if exc_type is not None:
            self.discard()
        super().__exit__(exc_type, exc, tb)

    def close(self) -> None:
        if not self.closed and not self._discard:
            _client().put_object(Bucket=self._bucket, Key=self._key, Body=self.getvalue())
        super().close()


class _TextWriter(io.TextIOWrapper):
    """Text-mode ``_Writer`` that also discards the buffer on an exception."""

    def __init__(self, writer: _Writer) -> None:
        super().__init__(writer, encoding="utf-8")
        self._writer = writer

    def __exit__(self, exc_type, exc, tb) -> None:
        if exc_type is not None:
            self._writer.discard()
        super().__exit__(exc_type, exc, tb)


def build_s3_filesystem(base_cls: type, not_found_error: type[Exception]) -> type:
    """Build the plugin class on TFX's ``Filesystem`` and ``NotFoundError``.

    Both are passed in so this module does not import TFX.
    """

    class S3Filesystem(base_cls):
        SUPPORTED_SCHEMES = [SCHEME]

        @staticmethod
        def open(name: PathType, mode: str = "r") -> IO:
            bucket, key = _split(name)
            if "a" in mode or "+" in mode:
                raise ValueError(f"Mode {mode!r} is not supported for {name!r}")
            if "w" in mode:
                writer = _Writer(bucket, key)
                return writer if "b" in mode else _TextWriter(writer)
            try:
                body = _client().get_object(Bucket=bucket, Key=key)["Body"].read()
            except _client_error() as exc:
                if _is_not_found(exc):
                    raise not_found_error(f"{name!r} not found") from exc
                raise
            return io.BytesIO(body) if "b" in mode else io.StringIO(body.decode("utf-8"))

        @staticmethod
        def copy(src: PathType, dst: PathType, overwrite: bool = False) -> None:
            if not overwrite and S3Filesystem.exists(dst):
                raise OSError(f"Destination file {dst!r} already exists")
            src_bucket, src_key = _split(src)
            dst_bucket, dst_key = _split(dst)
            if _head(src_bucket, src_key) is None:
                raise not_found_error(f"{src!r} not found")
            _client().copy_object(
                CopySource={"Bucket": src_bucket, "Key": src_key},
                Bucket=dst_bucket,
                Key=dst_key,
            )

        @staticmethod
        def exists(path: PathType) -> bool:
            bucket, key = _split(path)
            return _head(bucket, key) is not None or S3Filesystem.isdir(path)

        @staticmethod
        def glob(pattern: PathType) -> list[str]:
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
        def isdir(path: PathType) -> bool:
            bucket, key = _split(path)
            if not key:
                return True
            page = _client().list_objects_v2(Bucket=bucket, Prefix=_dir_prefix(key), MaxKeys=1)
            return bool(page.get("Contents") or page.get("CommonPrefixes"))

        @staticmethod
        def listdir(path: PathType) -> list[str]:
            bucket, key = _split(path)
            dirs, files = _children(bucket, key)
            if not dirs and not files and not S3Filesystem.isdir(path):
                raise not_found_error(f"{path!r} not found")
            return dirs + files

        @staticmethod
        def makedirs(path: PathType) -> None:
            _split(path)

        @staticmethod
        def mkdir(path: PathType) -> None:
            _split(path)

        @staticmethod
        def remove(path: PathType) -> None:
            bucket, key = _split(path)
            if _head(bucket, key) is None:
                raise not_found_error(f"{path!r} not found")
            _client().delete_object(Bucket=bucket, Key=key)

        @staticmethod
        def rename(src: PathType, dst: PathType, overwrite: bool = False) -> None:
            S3Filesystem.copy(src, dst, overwrite=overwrite)
            S3Filesystem.remove(src)

        @staticmethod
        def rmtree(path: PathType) -> None:
            bucket, key = _split(path)
            keys = list(_keys(bucket, _dir_prefix(key)))
            if not keys:
                raise not_found_error(f"{path!r} not found")
            for start in range(0, len(keys), _DELETE_BATCH_SIZE):
                batch = keys[start:start + _DELETE_BATCH_SIZE]
                _client().delete_objects(
                    Bucket=bucket,
                    Delete={"Objects": [{"Key": k} for k in batch]},
                )

        @staticmethod
        def stat(path: PathType) -> S3Stat:
            bucket, key = _split(path)
            head = _head(bucket, key)
            if head is not None:
                return S3Stat(
                    length=head["ContentLength"],
                    mtime_nsec=int(head["LastModified"].timestamp() * 1e9),
                    is_directory=False,
                )
            if S3Filesystem.isdir(path):
                return S3Stat(length=0, mtime_nsec=0, is_directory=True)
            raise not_found_error(f"{path!r} not found")

        @staticmethod
        def walk(
            top: PathType,
            topdown: bool = True,
            onerror: Callable[[Exception], None] | None = None,
        ) -> Iterator[tuple[str, list[str], list[str]]]:
            bucket, key = _split(top)
            top = f"{SCHEME}{bucket}/{key}".rstrip("/")
            try:
                dirs, files = _children(bucket, key)
            except _client_error() as exc:
                if onerror is not None:
                    onerror(exc)
                return
            if key and not dirs and not files:
                return
            if topdown:
                yield top, dirs, files
            for name in dirs:
                yield from S3Filesystem.walk(f"{top}/{name}", topdown, onerror)
            if not topdown:
                yield top, dirs, files

    return S3Filesystem
