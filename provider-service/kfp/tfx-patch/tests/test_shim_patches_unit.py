"""Patch tests with dependency-free fakes."""

import sys
import types

from tfx_kfp_v2_shim import _patches


class FakeType:
    def __init__(self, kind=None, instance_schema="", value=None):
        self.kind = kind
        self.instance_schema = instance_schema
        self.value = value

    def WhichOneof(self, name):
        assert name == "kind"
        return self.kind

    def CopyFrom(self, other):
        self.kind = other.kind
        self.instance_schema = other.instance_schema
        self.value = other.value


def test_patch_compiler_utils_adds_system_mappings(monkeypatch):
    model = type("Model", (), {})
    dataset = type("Dataset", (), {})
    metrics = type("Metrics", (), {})
    statistics = type("Statistics", (), {})
    generic = type("GenericArtifact", (), {})

    types_module = types.ModuleType("tfx.types")
    types_module.standard_artifacts = types.SimpleNamespace(Model=model)
    types_module.system_artifacts = types.SimpleNamespace(
        Dataset=dataset,
        Metrics=metrics,
        Statistics=statistics,
    )
    utils_module = types.ModuleType("tfx.utils")
    utils_module.name_utils = types.SimpleNamespace(get_full_name=lambda cls: f"fake.{cls.__name__}")
    tfx_module = types.ModuleType("tfx")
    tfx_module.types = types_module
    tfx_module.utils = utils_module
    monkeypatch.setitem(sys.modules, "tfx", tfx_module)
    monkeypatch.setitem(sys.modules, "tfx.types", types_module)
    monkeypatch.setitem(sys.modules, "tfx.utils", utils_module)
    monkeypatch.setattr(_patches, "_get_generic_artifact_class", lambda: generic)

    mod = types.SimpleNamespace(TITLE_TO_CLASS_PATH={})
    _patches.patch_compiler_utils(mod)

    assert mod.TITLE_TO_CLASS_PATH == {
        "system.Artifact": "fake.GenericArtifact",
        "system.Model": "fake.Model",
        "system.Dataset": "fake.Dataset",
        "system.Metrics": "fake.Metrics",
        "system.Statistics": "fake.Statistics",
    }


def test_enrich_input_artifacts_only_fills_untyped_matching_inputs():
    schema = FakeType("schema_title", value="tfx.Examples")
    missing_kind = types.SimpleNamespace(type=FakeType())
    empty_instance = types.SimpleNamespace(type=FakeType("instance_schema", ""))
    typed = types.SimpleNamespace(type=FakeType("schema_title", value="existing"))
    artifact_dict = {
        "input": types.SimpleNamespace(artifacts=[missing_kind, empty_instance, typed]),
        "other": types.SimpleNamespace(artifacts=[types.SimpleNamespace(type=FakeType())]),
    }
    inputs_spec = types.SimpleNamespace(artifacts={"input": types.SimpleNamespace(artifact_type=schema)})

    _patches._enrich_input_artifacts(artifact_dict, inputs_spec)

    assert missing_kind.type.value == "tfx.Examples"
    assert empty_instance.type.value == "tfx.Examples"
    assert typed.type.value == "existing"
    assert artifact_dict["other"].artifacts[0].type.kind is None


def test_parse_raw_artifact_dict_reads_inputs_spec_from_caller(monkeypatch):
    original_calls = []
    enriched = []

    def parse_raw_artifact_dict(artifact_dict, name):
        original_calls.append((artifact_dict, name))
        return "parsed"

    def enrich_input_artifacts(artifacts, spec):
        enriched.append((artifacts, spec))

    mod = types.SimpleNamespace(parse_raw_artifact_dict=parse_raw_artifact_dict)
    monkeypatch.setattr(_patches, "_enrich_input_artifacts", enrich_input_artifacts)
    _patches._patch_parse_raw_artifact_dict(mod)
    artifacts = {"key": "value"}
    inputs_spec = object()

    def caller_with_spec(inputs_spec):
        return mod.parse_raw_artifact_dict(artifacts, "with-spec")

    assert caller_with_spec(inputs_spec) == "parsed"
    assert enriched == [(artifacts, inputs_spec)]
    assert original_calls == [(artifacts, "with-spec")]


def test_patch_experimental_reexports_runner(monkeypatch):
    runner_module = types.ModuleType("tfx.orchestration.kubeflow.v2.kubeflow_v2_dag_runner")
    runner_module.KubeflowV2DagRunner = object
    runner_module.KubeflowV2DagRunnerConfig = object
    package_names = [
        "tfx",
        "tfx.orchestration",
        "tfx.orchestration.kubeflow",
        "tfx.orchestration.kubeflow.v2",
    ]
    fake_packages = {name: types.ModuleType(name) for name in package_names}
    v2_package = fake_packages["tfx.orchestration.kubeflow.v2"]
    v2_package.kubeflow_v2_dag_runner = runner_module
    modules = {**fake_packages, runner_module.__name__: runner_module}
    for name, module in modules.items():
        monkeypatch.setitem(sys.modules, name, module)
    mod = types.SimpleNamespace()
    _patches.patch_experimental(mod)
    assert mod.KubeflowV2DagRunner is runner_module.KubeflowV2DagRunner
    assert mod.KubeflowV2DagRunnerConfig is runner_module.KubeflowV2DagRunnerConfig


def test_patch_run_executor_exits_after_original(monkeypatch):
    calls = []
    exits = []

    def run_executor(*args, **kwargs):
        calls.append((args, kwargs))

    mod = types.SimpleNamespace(_run_executor=run_executor)
    monkeypatch.setattr(_patches.os, "_exit", lambda code: exits.append(code))
    _patches.patch_run_executor(mod)
    mod._run_executor("arg", option=True)
    assert calls == [(("arg",), {"option": True})]
    assert exits == [0]
