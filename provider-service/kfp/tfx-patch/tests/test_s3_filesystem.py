"""Unit tests for the boto3-backed s3:// fileio plugin (patch 10).

An in-memory fake stands in for the boto3 S3 client, so these tests run
without boto3, TFX or AWS access.
"""

from __future__ import annotations

import datetime
import io

import pytest

from tfx_kfp_v2_shim import _s3_filesystem


class FakeClientError(Exception):
    def __init__(self, code):
        super().__init__(code)
        self.response = {"Error": {"Code": code}}


class FakeS3Client:
    """Implements the subset of the boto3 S3 client the plugin uses."""

    def __init__(self, page_size=1000):
        self.objects: dict[tuple[str, str], bytes] = {}
        self.page_size = page_size

    def put_object(self, Bucket, Key, Body):
        self.objects[(Bucket, Key)] = bytes(Body)

    def get_object(self, Bucket, Key):
        if (Bucket, Key) not in self.objects:
            raise FakeClientError("NoSuchKey")
        return {"Body": io.BytesIO(self.objects[(Bucket, Key)])}

    def head_object(self, Bucket, Key):
        if (Bucket, Key) not in self.objects:
            raise FakeClientError("404")
        return {
            "ContentLength": len(self.objects[(Bucket, Key)]),
            "LastModified": datetime.datetime(2026, 1, 1, tzinfo=datetime.timezone.utc),
        }

    def copy_object(self, CopySource, Bucket, Key):
        self.objects[(Bucket, Key)] = self.objects[(CopySource["Bucket"], CopySource["Key"])]

    def delete_object(self, Bucket, Key):
        self.objects.pop((Bucket, Key), None)

    def delete_objects(self, Bucket, Delete):
        for obj in Delete["Objects"]:
            self.objects.pop((Bucket, obj["Key"]), None)

    def list_objects_v2(self, Bucket, Prefix, Delimiter=None, MaxKeys=None,
                        ContinuationToken=None):
        keys = sorted(k for b, k in self.objects if b == Bucket and k.startswith(Prefix))
        contents, prefixes = [], []
        for key in keys:
            rest = key[len(Prefix):]
            if Delimiter and Delimiter in rest:
                common = Prefix + rest.split(Delimiter, 1)[0] + Delimiter
                if common not in prefixes:
                    prefixes.append(common)
            else:
                contents.append({"Key": key})

        entries = [("c", c) for c in contents] + [("p", p) for p in prefixes]
        start = int(ContinuationToken or 0)
        size = min(self.page_size, MaxKeys or self.page_size)
        page_entries = entries[start:start + size]
        page = {
            "Contents": [e for kind, e in page_entries if kind == "c"],
            "CommonPrefixes": [{"Prefix": e} for kind, e in page_entries if kind == "p"],
            "IsTruncated": start + size < len(entries),
        }
        if page["IsTruncated"]:
            page["NextContinuationToken"] = str(start + size)
        return page


class FakeBase:
    pass


class FakeNotFoundError(IOError):
    pass


@pytest.fixture
def client(monkeypatch):
    fake = FakeS3Client(page_size=2)
    monkeypatch.setattr(_s3_filesystem, "_client", lambda: fake)
    return fake


@pytest.fixture
def fs(client):
    return _s3_filesystem.build_s3_filesystem(FakeBase, FakeNotFoundError)


def test_claims_only_s3_scheme_ahead_of_tensorflow(fs):
    assert fs.SUPPORTED_SCHEMES == ["s3://"]
    assert issubclass(fs, FakeBase)
    assert _s3_filesystem.PRIORITY < 10


def test_write_then_read_binary_and_text(fs, client):
    with fs.open("s3://bucket/dir/file.bin", "wb") as f:
        f.write(b"\x00\x01")
    with fs.open("s3://bucket/dir/file.txt", "w") as f:
        f.write("hello")

    assert client.objects[("bucket", "dir/file.bin")] == b"\x00\x01"
    with fs.open("s3://bucket/dir/file.bin", "rb") as f:
        assert f.read() == b"\x00\x01"
    with fs.open(b"s3://bucket/dir/file.txt") as f:
        assert f.read() == "hello"


def test_open_missing_raises_not_found(fs):
    with pytest.raises(FakeNotFoundError):
        fs.open("s3://bucket/missing", "rb")


def test_open_append_unsupported(fs):
    with pytest.raises(ValueError):
        fs.open("s3://bucket/file", "a")


def test_non_s3_path_rejected(fs):
    with pytest.raises(ValueError):
        fs.exists("/local/path")


def test_exists_and_isdir(fs, client):
    client.put_object("bucket", "dir/sub/file", b"x")

    assert fs.exists("s3://bucket/dir/sub/file")
    assert not fs.isdir("s3://bucket/dir/sub/file")
    assert fs.exists("s3://bucket/dir")
    assert fs.isdir("s3://bucket/dir/")
    assert fs.isdir("s3://bucket")
    assert not fs.exists("s3://bucket/other")


def test_listdir_paginates(fs, client):
    for key in ("dir/a", "dir/b", "dir/c", "dir/sub/d", "dir/sub2/e"):
        client.put_object("bucket", key, b"x")

    assert sorted(fs.listdir("s3://bucket/dir")) == ["a", "b", "c", "sub", "sub2"]
    with pytest.raises(FakeNotFoundError):
        fs.listdir("s3://bucket/missing")


def test_walk_topdown_and_bottomup(fs, client):
    for key in ("root/a", "root/sub/b", "root/sub/deeper/c"):
        client.put_object("bucket", key, b"x")

    assert list(fs.walk("s3://bucket/root/")) == [
        ("s3://bucket/root", ["sub"], ["a"]),
        ("s3://bucket/root/sub", ["deeper"], ["b"]),
        ("s3://bucket/root/sub/deeper", [], ["c"]),
    ]
    assert [d for d, _, _ in fs.walk("s3://bucket/root", topdown=False)] == [
        "s3://bucket/root/sub/deeper",
        "s3://bucket/root/sub",
        "s3://bucket/root",
    ]
    assert list(fs.walk("s3://bucket/missing")) == []


def test_glob(fs, client):
    for key in ("dir/a.txt", "dir/b.txt", "dir/c.csv"):
        client.put_object("bucket", key, b"x")

    assert fs.glob("s3://bucket/dir/*.txt") == ["s3://bucket/dir/a.txt", "s3://bucket/dir/b.txt"]


def test_copy_respects_overwrite(fs, client):
    client.put_object("bucket", "src", b"1")
    client.put_object("bucket", "dst", b"2")

    with pytest.raises(OSError):
        fs.copy("s3://bucket/src", "s3://bucket/dst")
    fs.copy("s3://bucket/src", "s3://bucket/dst", overwrite=True)
    assert client.objects[("bucket", "dst")] == b"1"
    with pytest.raises(FakeNotFoundError):
        fs.copy("s3://bucket/missing", "s3://bucket/new")


def test_rename(fs, client):
    client.put_object("bucket", "src", b"1")

    fs.rename("s3://bucket/src", "s3://other/dst")
    assert client.objects == {("other", "dst"): b"1"}


def test_remove(fs, client):
    client.put_object("bucket", "file", b"1")

    fs.remove("s3://bucket/file")
    assert client.objects == {}
    with pytest.raises(FakeNotFoundError):
        fs.remove("s3://bucket/file")


def test_rmtree_deletes_only_prefix(fs, client):
    for key in ("dir/a", "dir/sub/b", "dir2/c"):
        client.put_object("bucket", key, b"x")

    fs.rmtree("s3://bucket/dir")
    assert list(client.objects) == [("bucket", "dir2/c")]
    with pytest.raises(FakeNotFoundError):
        fs.rmtree("s3://bucket/dir")


def test_makedirs_is_noop(fs, client):
    fs.makedirs("s3://bucket/new/dir")
    fs.mkdir("s3://bucket/new")
    assert client.objects == {}


def test_stat(fs, client):
    client.put_object("bucket", "dir/file", b"abc")

    file_stat = fs.stat("s3://bucket/dir/file")
    assert file_stat.length == 3
    assert not file_stat.is_directory
    assert file_stat.mtime_nsec > 0
    assert fs.stat("s3://bucket/dir").is_directory
    with pytest.raises(FakeNotFoundError):
        fs.stat("s3://bucket/missing")
