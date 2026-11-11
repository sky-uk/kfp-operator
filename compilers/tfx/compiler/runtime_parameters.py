import contextlib
import functools
from collections.abc import Iterator
from typing import Any

from kfp.pipeline_spec import pipeline_spec_pb2
from tfx.orchestration import data_types
from tfx.orchestration.kubeflow.v2 import compiler_utils

ParameterSpec = pipeline_spec_pb2.ComponentInputsSpec.ParameterSpec

_NUMBER_TYPES = (
    pipeline_spec_pb2.ParameterType.NUMBER_INTEGER,
    pipeline_spec_pb2.ParameterType.NUMBER_DOUBLE,
)


def _apply_runtime_parameter_default(parameter_spec: ParameterSpec, value: Any) -> None:
    """Copy a TFX RuntimeParameter default onto a KFP ParameterSpec.

    TFX records the default on RuntimeParameter and then drops it when building
    the spec. A missing default makes the parameter required.
    """
    if not isinstance(value, data_types.RuntimeParameter) or value.default is None:
        return

    default = value.default
    parameter_type = parameter_spec.parameter_type
    if parameter_type in _NUMBER_TYPES and isinstance(default, (int, float)):
        parameter_spec.default_value.number_value = default
    elif parameter_type == pipeline_spec_pb2.ParameterType.STRING and isinstance(default, (str, int, float)):
        parameter_spec.default_value.string_value = str(default)
    else:
        return

    # kfp-pipeline-spec 0.1.14 has no is_optional; KFP still applies the default without it.
    if "is_optional" in parameter_spec.DESCRIPTOR.fields_by_name:
        parameter_spec.is_optional = True


@contextlib.contextmanager
def runtime_parameter_defaults() -> Iterator[None]:
    """Keep RuntimeParameter defaults while TFX builds a pipeline spec 2.1."""
    original = compiler_utils.build_parameter_type_spec

    @functools.wraps(original)
    def build(value: Any) -> ParameterSpec:
        parameter_spec = original(value)
        _apply_runtime_parameter_default(parameter_spec, value)
        return parameter_spec

    compiler_utils.build_parameter_type_spec = build
    try:
        yield
    finally:
        compiler_utils.build_parameter_type_spec = original
