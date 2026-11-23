"""Artifact type and schema tests for tfx_kfp_v2_shim."""

import unittest

from tfx_kfp_v2_shim import _patches

try:
    import yaml  # noqa: F401

    _HAS_YAML = True
except Exception:  # pragma: no cover
    _HAS_YAML = False

try:
    from tfx.types import artifact as _artifact_mod
    from tfx.types import standard_artifacts

    _HAS_TFX = True
except Exception:  # pragma: no cover
    _HAS_TFX = False


# ── Fake KFP-v2 RuntimeArtifact protos (enough surface for the patch) ────────


class _FakeType:
    def __init__(self, schema_title=None, instance_schema=None):
        self.schema_title = schema_title
        self.instance_schema = instance_schema

    def WhichOneof(self, oneof):
        assert oneof == "kind"
        if self.schema_title is not None:
            return "schema_title"
        if self.instance_schema is not None:
            return "instance_schema"
        return None


class _FakeMetadata:
    def __init__(self, fields):
        self.fields = dict(fields)


class _FakeArtifactPb:
    def __init__(self, type_, keys=()):
        self.type = type_
        self.metadata = _FakeMetadata({k: None for k in keys})
        self.properties = {}


class _FakeMod:
    """Stands in for kubeflow_v2_entrypoint_utils; records the wrapped call."""

    def __init__(self):
        self.calls = []

        def _orig(artifact_pb, name_from_id):
            self.calls.append((artifact_pb, name_from_id))
            return "ORIGINAL"

        self._parse_raw_artifact = _orig


def _parse(schema_title=None, instance_schema=None, keys=()):
    """Apply the patch to a fresh fake module and run one artifact through it."""
    mod = _FakeMod()
    _patches._patch_parse_raw_artifact(mod)
    pb = _FakeArtifactPb(_FakeType(schema_title, instance_schema), keys)
    result = mod._parse_raw_artifact(pb, None)
    return pb, result, mod


@unittest.skipUnless(_HAS_YAML, "PyYAML required")
class RelabelBehaviorTest(unittest.TestCase):
    """Artifact disambiguation and null-schema handling per (title, keys)."""

    def test_model_with_pushed_destination_becomes_pushed_model(self):
        pb, result, _ = _parse("system.Model", keys={"pushed_destination"})
        self.assertEqual(pb.type.schema_title, "tfx.PushedModel")
        self.assertEqual(result, "ORIGINAL")

    def test_model_with_pushed_becomes_pushed_model(self):
        pb, _, _ = _parse("system.Model", keys={"pushed"})
        self.assertEqual(pb.type.schema_title, "tfx.PushedModel")

    def test_model_without_keys_stays_model(self):
        pb, _, _ = _parse("system.Model", keys=set())
        self.assertEqual(pb.type.schema_title, "system.Model")

    def test_dataset_with_split_names_becomes_examples(self):
        pb, _, _ = _parse("system.Dataset", keys={"split_names"})
        self.assertEqual(pb.type.schema_title, "tfx.Examples")

    def test_dataset_without_split_names_stays_dataset(self):
        pb, _, _ = _parse("system.Dataset", keys={"span", "version"})
        self.assertEqual(pb.type.schema_title, "system.Dataset")

    def test_statistics_has_no_rule_and_is_untouched(self):
        pb, _, _ = _parse("system.Statistics", keys={"split_names"})
        self.assertEqual(pb.type.schema_title, "system.Statistics")

    def test_subset_match_ignores_extra_keys(self):
        pb, _, _ = _parse("system.Model", keys={"pushed_destination", "span", "custom"})
        self.assertEqual(pb.type.schema_title, "tfx.PushedModel")

    def test_empty_instance_schema_is_guarded(self):
        # yaml.safe_load("") is None; the guard must not raise and must backfill
        # a minimal title so downstream parsing succeeds.
        pb, result, _ = _parse(instance_schema="")
        self.assertEqual(pb.type.instance_schema, "title: ''")
        self.assertEqual(result, "ORIGINAL")


def _system_title(art_cls):
    """Runtime system.* title the KFP-v2 driver assigns from TYPE_ANNOTATION."""
    ann = getattr(art_cls, "TYPE_ANNOTATION", None)
    return "system." + ann.__name__ if ann is not None else "system.Artifact"


def _standard_tfx_types():
    for name in dir(standard_artifacts):
        obj = getattr(standard_artifacts, name)
        if (
            isinstance(obj, type)
            and issubclass(obj, _artifact_mod.Artifact)
            and obj is not _artifact_mod.Artifact
            and getattr(obj, "TYPE_NAME", None)
        ):
            yield obj


@unittest.skipUnless(_HAS_TFX, "tfx required (linux/x86_64 only)")
class CompletenessTest(unittest.TestCase):
    """Verify collision rules against the installed TFX artifact types."""

    def test_only_known_collision_is_covered_by_a_rule(self):
        # Group every producible standard type by its runtime system.* title;
        # a title claimed by >1 distinct type is a collision the shim must
        # disambiguate. A future TFX bump adding a new collision fails here.
        by_title = {}
        for cls in _standard_tfx_types():
            by_title.setdefault(_system_title(cls), set()).add(cls.TYPE_NAME)

        # system.Artifact is the shared generic bucket (Schema, ModelBlessing,
        # value artifacts, ...); it is disambiguated later from inputs_spec
        # (input artifact schema enrichment), not by a rule, so it is excluded
        # from the collision check.
        collisions = {t: n for t, n in by_title.items() if len(n) > 1 and t != "system.Artifact"}
        self.assertEqual(
            collisions,
            {"system.Model": {"Model", "PushedModel"}},
            "New/changed system.* collision — add a TYPE_INFERENCE_RULES entry.",
        )

        rules = _patches.TYPE_INFERENCE_RULES["system.Model"]
        targets = {tfx_title for _, tfx_title in rules}
        self.assertIn("tfx.PushedModel", targets)

    def test_examples_rule_is_statically_reachable(self):
        # Examples.split_names is a declared PROPERTY, so its rule is reachable
        # from compile-time schema alone. PushedModel's pushed/pushed_destination
        # keys are covered by the behavior tests above.
        rules = _patches.TYPE_INFERENCE_RULES["system.Dataset"]
        signatures = [sig for sig, _ in rules]
        self.assertTrue(
            any(sig <= set(standard_artifacts.Examples.PROPERTIES) for sig in signatures),
            "Examples disambiguation key not found in its PROPERTIES.",
        )

    def test_compiler_mappings_cover_every_producible_title(self):
        # The compiler mappings ensure no producible system.* title raises
        # ImportError.
        class _Mod:
            TITLE_TO_CLASS_PATH = {}

        mod = _Mod()
        _patches.patch_compiler_utils(mod)
        produced = {_system_title(cls) for cls in _standard_tfx_types()}
        missing = produced - set(mod.TITLE_TO_CLASS_PATH)
        self.assertEqual(missing, set(), f"Titles unmapped by compiler mappings: {missing}")


if __name__ == "__main__":
    unittest.main()
