"""Unit tests for the TFX KFP v2 import hook and shim patches.

These tests use mock modules and lightweight stub objects so they can run
without TFX or KFP installed.
"""

from __future__ import annotations

import sys
import types
from unittest.mock import MagicMock

import pytest

from tfx_kfp_v2_shim._hook import TfxShimFinder
from tfx_kfp_v2_shim._patches import (
    TYPE_INFERENCE_RULES,
    _collect_artifact_keys,
    _enrich_input_artifacts,
    patch_compiler_utils,
    patch_entrypoint_utils,
    patch_path_utils,
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

    def artifacts(self):
        pass  # replaced below

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


# ── Tests: Patch 1 (compiler_utils) ─────────────────────────────────────


class TestPatchCompilerUtils:
    def test_adds_system_mappings(self):
        """patch_compiler_utils should add system.* entries to TITLE_TO_CLASS_PATH."""
        # Create a fake compiler_utils module with the required TFX dependencies
        mod = types.ModuleType("fake_compiler_utils")
        mod.TITLE_TO_CLASS_PATH = {"tfx.Model": "tfx.types.standard_artifacts.Model"}

        # We need the real TFX imports to work; if TFX is not installed, skip.
        try:
            from tfx.types import simple_artifacts, standard_artifacts
            from tfx.utils import import_utils, name_utils
        except ImportError:
            pytest.skip("TFX not installed")

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


# ── Tests: Patch 2+3 (_parse_raw_artifact wrapper) ──────────────────────


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
        """Patch 3: system.Dataset with split_names → tfx.Examples."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(
            FakeTypeSchema(schema_title="system.Dataset"),
            metadata=FakeMetadata({"split_names": "train", "span": 1}),
        )
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "tfx.Examples"

    def test_infers_pushed_model_from_metadata(self):
        """Patch 3: system.Model with pushed → tfx.PushedModel."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(
            FakeTypeSchema(schema_title="system.Model"),
            properties={"pushed": 1, "pushed_destination": "gs://..."},
        )
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "tfx.PushedModel"

    def test_no_inference_for_unknown_title(self):
        """Patch 3: unknown titles are left alone."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(schema_title="system.Foo"))
        mod._parse_raw_artifact(art, {})
        assert art.type.schema_title == "system.Foo"

    def test_null_instance_schema_guarded(self):
        """Patch 2: empty instance_schema gets a safe default."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(instance_schema=""))
        mod._parse_raw_artifact(art, {})
        assert art.type.instance_schema == "title: ''"

    def test_valid_instance_schema_unchanged(self):
        """Patch 2: non-empty instance_schema is not modified."""
        mod = self._make_module_with_parser()
        patch_entrypoint_utils(mod)

        art = FakeArtifact(FakeTypeSchema(instance_schema="title: tfx.Model"))
        mod._parse_raw_artifact(art, {})
        assert art.type.instance_schema == "title: tfx.Model"


# ── Tests: Patch 4 (parse_raw_artifact_dict enrichment) ─────────────────


class TestPatchParseRawArtifactDict:
    def test_enriches_untyped_artifacts(self):
        """Patch 4: artifacts without type get it from inputs_spec."""
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
        """Patch 4: no enrichment when inputs_spec is not in caller's scope."""
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


# ── Tests: Patch 5 (path_utils) ─────────────────────────────────────────


class TestPatchPathUtils:
    def test_flattens_serving_dir(self):
        """patch_path_utils should make serving_model_dir return output_uri."""
        mod = types.ModuleType("fake_path_utils")
        mod.serving_model_dir = lambda output_uri: output_uri + "/Format-Serving"
        patch_path_utils(mod)

        assert mod.serving_model_dir("gs://bucket/model") == "gs://bucket/model"


# ── Tests: Patch 8 (run_executor force-exit) ─────────────────────────────


class TestPatchRunExecutor:
    def test_forces_exit_after_run(self, monkeypatch):
        """Patch 8: _run_executor should call os._exit(0) after completing."""
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
        """Patch 8: the original _run_executor must run before force-exit."""
        import os as _os

        order = []
        mod = types.ModuleType("fake_run_executor")
        mod._run_executor = lambda *a, **k: order.append("ran")
        patch_run_executor(mod)

        monkeypatch.setattr(_os, "_exit", lambda code: order.append(f"exit:{code}"))

        mod._run_executor("args", ["beam"])

        assert order == ["ran", "exit:0"]


# ── Tests: Import hook mechanism ─────────────────────────────────────────


class TestTfxShimFinder:
    def test_find_spec_returns_spec_for_target(self):
        finder = TfxShimFinder({"some.module": lambda m: None})
        spec = finder.find_spec("some.module", None)
        assert spec is not None
        assert spec.name == "some.module"

    def test_find_spec_returns_none_for_non_target(self):
        finder = TfxShimFinder({"some.module": lambda m: None})
        assert finder.find_spec("other.module", None) is None

    def test_register_and_unregister(self):
        finder = TfxShimFinder({})
        finder.register()
        assert finder in sys.meta_path
        finder.unregister()
        assert finder not in sys.meta_path

    def test_idempotent_register(self):
        finder = TfxShimFinder({})
        finder.register()
        finder.register()
        count = sys.meta_path.count(finder)
        finder.unregister()
        assert count == 1

    def test_patches_module_on_import(self):
        """The hook should intercept and patch a real module on import."""
        patched_flag = {}

        def my_patch(mod):
            patched_flag["called"] = True
            mod._SHIM_PATCHED = True

        # Use a real stdlib module (e.g. `colorsys`) as the target
        target = "colorsys"
        sys.modules.pop(target, None)

        finder = TfxShimFinder({target: my_patch})
        finder.register()
        try:
            import importlib
            mod = importlib.import_module(target)
            assert patched_flag.get("called") is True
            assert getattr(mod, "_SHIM_PATCHED", False) is True
        finally:
            finder.unregister()
            sys.modules.pop(target, None)
