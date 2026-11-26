"""Unit tests for the TFX KFP v2 shim patches.

These tests use mock modules and lightweight stub objects so they can run
without TFX or KFP installed.
"""

from __future__ import annotations

import sys
import types
from unittest.mock import MagicMock

import pytest

from tfx_kfp_v2_shim._patches import (
    _collect_artifact_keys,
    _delete_gcs_directory_markers,
    _patch_translate_executor_output,
    patch_compiler_utils,
    patch_entrypoint_utils,
    patch_run_executor,
)

# ── Helpers for building mock protobuf-like objects ───────────────────────


class FakeTypeSchema:
    """Mimics pipeline_pb2.ArtifactTypeSchema with a oneof ``kind``."""

    def __init__(self, *, schema_title=None, instance_schema=None):
        self.schema_title = schema_title
        self.instance_schema = instance_schema

    def WhichOneof(self, name):
        if name == "kind":
            if self.schema_title is not None:
                return "schema_title"
            if self.instance_schema is not None:
                return "instance_schema"
            return None
        raise ValueError(f"Unknown oneof: {name}")

    def CopyFrom(self, other):
        self.schema_title = other.schema_title
        self.instance_schema = other.instance_schema


class FakeMetadata:
    def __init__(self, fields: dict | None = None):
        self.fields = fields or {}


class FakeArtifact:
    """Mimics pipeline_pb2.RuntimeArtifact."""

    def __init__(self, type_schema, metadata=None, properties=None):
        self.type = type_schema
        self.metadata = metadata or FakeMetadata()
        self.properties = properties or {}


class FakeArtifactList:
    def __init__(self, artifacts):
        self.artifacts = artifacts


class FakeArtifactSpec:
    def __init__(self, artifact_type):
        self.artifact_type = artifact_type


class FakeInputsSpec:
    def __init__(self, artifacts: dict):
        self._artifacts = artifacts

    class _Items:
        def __init__(self, d):
            self._d = d

        def items(self):
            return self._d.items()

        def keys(self):
            return self._d.keys()

    @property
    def artifacts(self):
        return self._Items(self._artifacts)


# ── Tests: _collect_artifact_keys ────────────────────────────────────────


class TestCollectArtifactKeys:
    def test_empty(self):
        art = FakeArtifact(FakeTypeSchema())
        assert _collect_artifact_keys(art) == set()

    def test_metadata_keys(self):
        art = FakeArtifact(
            FakeTypeSchema(),
            metadata=FakeMetadata({"split_names": 1, "span": 2}),
        )
        assert _collect_artifact_keys(art) == {"split_names", "span"}

    def test_property_keys(self):
        art = FakeArtifact(FakeTypeSchema(), properties={"pushed": 1})
        assert _collect_artifact_keys(art) == {"pushed"}

    def test_combined(self):
        art = FakeArtifact(
            FakeTypeSchema(),
            metadata=FakeMetadata({"a": 1}),
            properties={"b": 2},
        )
        assert _collect_artifact_keys(art) == {"a", "b"}


# ── Tests: artifact type mappings (patch_compiler_utils) ─────────────────────────────────────


class TestPatchCompilerUtils:
    def test_adds_system_mappings(self):
        """patch_compiler_utils should add system.* entries to TITLE_TO_CLASS_PATH."""
        # Create a fake compiler_utils module with the required TFX dependencies
        mod = types.ModuleType("fake_compiler_utils")
        mod.TITLE_TO_CLASS_PATH = {"tfx.Model": "tfx.types.standard_artifacts.Model"}

        # We need the real TFX imports to work; if TFX is not installed, skip.
        pytest.importorskip("tfx")
        from tfx.utils import import_utils

        patch_compiler_utils(mod)

        assert "system.Model" in mod.TITLE_TO_CLASS_PATH
        assert "system.Dataset" in mod.TITLE_TO_CLASS_PATH
        assert "system.Artifact" in mod.TITLE_TO_CLASS_PATH
        assert "system.Metrics" in mod.TITLE_TO_CLASS_PATH
        assert "system.Statistics" in mod.TITLE_TO_CLASS_PATH
        # Original entry preserved
        assert "tfx.Model" in mod.TITLE_TO_CLASS_PATH

        # Regression: every stored path must round-trip through TFX's import
        # machinery. system.Artifact previously mapped to a dynamically created
        # class that was not bound to its declared module, so import_class_by_path
        # raised "not importable" at compile time.
        for title in ("system.Artifact", "system.Model"):
            import_utils.import_class_by_path(mod.TITLE_TO_CLASS_PATH[title])


# ── Tests: null-schema handling and type inference (_parse_raw_artifact wrapper) ──────────────────────


class TestPatchParseRawArtifact:
    def _make_module_with_parser(self):
        """Create a mock entrypoint_utils module with _parse_raw_artifact."""
        mod = types.ModuleType("fake_entrypoint_utils")
        mod._parse_raw_artifact = MagicMock(return_value="parsed")
        mod.parse_raw_artifact_dict = MagicMock(return_value={})
        mod.translate_executor_output = MagicMock(return_value={})
        return mod

    def test_delegates_to_original(self):
        mod = self._make_module_with_parser()
        original = mod._parse_raw_artifact
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(schema_title="tfx.Foo"))
        result = mod._parse_raw_artifact(art, {})

        assert result == "parsed"
        original.assert_called_once()

    def test_infers_examples_from_metadata(self):
        """Type inference: system.Dataset with split_names → tfx.Examples."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(
            FakeTypeSchema(schema_title="system.Dataset"),
            metadata=FakeMetadata({"split_names": "train", "span": 1}),
        )
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "tfx.Examples"

    def test_infers_pushed_model_from_metadata(self):
        """Type inference: system.Model with pushed → tfx.PushedModel."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(
            FakeTypeSchema(schema_title="system.Model"),
            properties={"pushed": 1, "pushed_destination": "gs://..."},
        )
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "tfx.PushedModel"

    def test_no_inference_for_unknown_title(self):
        """Type inference: unknown titles are left alone."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(schema_title="system.Foo"))
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "system.Foo"

    def test_null_instance_schema_guarded(self):
        """Null-schema handling: empty instance_schema gets a safe default."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(instance_schema=""))
        mod._parse_raw_artifact(art, {})
        assert art.type.instance_schema == "title: ''"

    def test_valid_instance_schema_unchanged(self):
        """Null-schema handling: non-empty instance_schema is not modified."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(instance_schema="title: tfx.Model"))
        mod._parse_raw_artifact(art, {})
        assert art.type.instance_schema == "title: tfx.Model"


# ── Tests: input artifact schema enrichment (parse_raw_artifact_dict) ─────────────────


class TestPatchParseRawArtifactDict:
    def test_enriches_untyped_artifacts(self):
        """Schema enrichment: artifacts without type get it from inputs_spec."""
        mod = types.ModuleType("fake_entrypoint_utils")

        def mock_parse(artifact_dict, name_from_id):
            return {}

        mod._parse_raw_artifact = MagicMock(return_value="parsed")
        mod.parse_raw_artifact_dict = mock_parse
        mod.translate_executor_output = MagicMock(return_value={})
        patch_entrypoint_utils(mod)

        # Build test data
        spec_type = FakeTypeSchema(schema_title="tfx.Examples")
        art = FakeArtifact(FakeTypeSchema())  # no type set
        artifact_dict = {"examples": FakeArtifactList([art])}
        inputs_spec = FakeInputsSpec({"examples": FakeArtifactSpec(spec_type)})  # noqa: F841

        # Call — the wrapper uses frame inspection to find inputs_spec
        mod.parse_raw_artifact_dict(artifact_dict, {})

        # The artifact should have been enriched
        assert art.type.schema_title == spec_type.schema_title

    def test_skips_enrichment_when_no_inputs_spec(self):
        """Schema enrichment: no enrichment when inputs_spec is not in caller's scope."""
        mod = types.ModuleType("fake_entrypoint_utils")
        mod._parse_raw_artifact = MagicMock(return_value="parsed")
        mod.parse_raw_artifact_dict = MagicMock(return_value={})
        mod.translate_executor_output = MagicMock(return_value={})
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema())
        artifact_dict = {"examples": FakeArtifactList([art])}

        # No inputs_spec in this scope
        mod.parse_raw_artifact_dict(artifact_dict, {})

        # Artifact type should remain unset
        assert art.type.schema_title is None


# ── Tests: executor shutdown compatibility (patch_run_executor) ─────────────────────────────


class TestPatchRunExecutor:
    def test_forces_exit_after_run(self, monkeypatch):
        """Executor shutdown: _run_executor should call os._exit(0) after completing."""
        import os as _os

        mod = types.ModuleType("fake_run_executor")
        original = MagicMock(return_value=None)
        mod._run_executor = original
        patch_run_executor(mod)

        exit_calls = []
        monkeypatch.setattr(_os, "_exit", lambda code: exit_calls.append(code))

        mod._run_executor("args", ["beam"])

        original.assert_called_once_with("args", ["beam"])
        assert exit_calls == [0]

    def test_runs_original_before_exit(self, monkeypatch):
        """Executor shutdown: the original _run_executor must run before force-exit."""
        import os as _os

        order = []
        mod = types.ModuleType("fake_run_executor")
        mod._run_executor = lambda *a, **k: order.append("ran")
        patch_run_executor(mod)

        monkeypatch.setattr(_os, "_exit", lambda code: order.append(f"exit:{code}"))

        mod._run_executor("args", ["beam"])

        assert order == ["ran", "exit:0"]


# ── Tests: GCS directory-marker cleanup ──────────────────────────────────


class FakeBlob:
    def __init__(self, name, size):
        self.name = name
        self.size = size
        self.deleted = False

    def delete(self):
        self.deleted = True


class FakeBucket:
    def __init__(self, blobs):
        self._blobs = blobs
        self.prefixes = []

    def list_blobs(self, prefix):
        self.prefixes.append(prefix)
        return self._blobs


class FakeGcsClient:
    def __init__(self, blobs):
        self.bucket_obj = FakeBucket(blobs)
        self.bucket_names = []
        self.batched = False

    def bucket(self, name):
        self.bucket_names.append(name)
        return self.bucket_obj

    def batch(self):
        client = self

        class _Batch:
            def __enter__(self):
                client.batched = True

            def __exit__(self, *exc):
                return False

        return _Batch()


class TestDeleteGcsDirectoryMarkers:
    def test_deletes_only_zero_byte_directory_markers(self):
        marker = FakeBlob("out/model/", 0)
        empty_file = FakeBlob("out/model/empty.txt", 0)
        non_empty_dir_named_blob = FakeBlob("out/model/dir/", 10)
        real_file = FakeBlob("out/model/saved_model.pb", 1234)
        client = FakeGcsClient([marker, empty_file, non_empty_dir_named_blob, real_file])

        _delete_gcs_directory_markers("gs://bucket/out/model", lambda: client)

        assert marker.deleted
        assert not empty_file.deleted
        assert not non_empty_dir_named_blob.deleted
        assert not real_file.deleted
        assert client.batched

    def test_splits_bucket_and_normalises_prefix(self):
        client = FakeGcsClient([])

        _delete_gcs_directory_markers("gs://my-bucket/path/to/artifact", lambda: client)

        assert client.bucket_names == ["my-bucket"]
        assert client.bucket_obj.prefixes == ["path/to/artifact/"]

    def test_no_batch_when_nothing_to_delete(self):
        client = FakeGcsClient([FakeBlob("out/file.txt", 5)])

        _delete_gcs_directory_markers("gs://bucket/out", lambda: client)

        assert not client.batched

    def test_swallows_errors(self):
        class ExplodingClient:
            def bucket(self, name):
                raise RuntimeError("gcs unavailable")

        _delete_gcs_directory_markers("gs://bucket/out", lambda: ExplodingClient())

    def test_swallows_client_creation_errors(self):
        def failing_factory():
            raise ImportError("google-cloud-storage is not installed")

        _delete_gcs_directory_markers("gs://bucket/out", failing_factory)


class TestPatchTranslateExecutorOutput:
    def _fake_gcs_module(self, monkeypatch, client_factory):
        storage = types.ModuleType("google.cloud.storage")
        storage.Client = client_factory
        cloud = types.ModuleType("google.cloud")
        cloud.storage = storage
        google = types.ModuleType("google")
        google.cloud = cloud
        monkeypatch.setitem(sys.modules, "google", google)
        monkeypatch.setitem(sys.modules, "google.cloud", cloud)
        monkeypatch.setitem(sys.modules, "google.cloud.storage", storage)

    def test_cleans_gcs_outputs_then_delegates(self, monkeypatch):
        marker = FakeBlob("out/model/", 0)
        client = FakeGcsClient([marker])
        self._fake_gcs_module(monkeypatch, lambda: client)

        mod = types.ModuleType("fake_entrypoint_utils")
        mod.translate_executor_output = MagicMock(return_value="translated")
        _patch_translate_executor_output(mod)

        gcs_artifact = types.SimpleNamespace(uri="gs://bucket/out/model")
        output_dict = {"model": [gcs_artifact]}

        assert mod.translate_executor_output(output_dict, {}) == "translated"
        assert marker.deleted

    def test_ignores_non_gcs_outputs(self, monkeypatch):
        client = FakeGcsClient([FakeBlob("out/model/", 0)])
        self._fake_gcs_module(monkeypatch, lambda: client)

        mod = types.ModuleType("fake_entrypoint_utils")
        mod.translate_executor_output = MagicMock(return_value="translated")
        _patch_translate_executor_output(mod)

        output_dict = {
            "model": [
                types.SimpleNamespace(uri="s3://bucket/out/model"),
                types.SimpleNamespace(uri=""),
            ]
        }

        assert mod.translate_executor_output(output_dict, {}) == "translated"
        assert client.bucket_names == []

    def test_delegates_when_gcs_client_cannot_be_created(self, monkeypatch):
        def failing_client():
            raise RuntimeError("no default credentials")

        self._fake_gcs_module(monkeypatch, failing_client)

        mod = types.ModuleType("fake_entrypoint_utils")
        mod.translate_executor_output = MagicMock(return_value="translated")
        _patch_translate_executor_output(mod)

        output_dict = {"model": [types.SimpleNamespace(uri="gs://bucket/out/model")]}

        assert mod.translate_executor_output(output_dict, {}) == "translated"
