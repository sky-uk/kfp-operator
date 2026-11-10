import pytest
from tfx.orchestration import data_types
from tfx.orchestration.kubeflow.v2 import compiler_utils

from compiler.runtime_parameters import runtime_parameter_defaults


@pytest.fixture
def build_parameter_type_spec():
    with runtime_parameter_defaults():
        yield compiler_utils.build_parameter_type_spec


@pytest.mark.parametrize("ptype, default, kind, expected", [
    (str, "us", "string_value", "us"),
    (str, "", "string_value", ""),
    (int, 0, "number_value", 0),
    (int, 5, "number_value", 5),
    (float, 0.5, "number_value", 0.5),
    (int, True, "number_value", 1),
    (None, 5, "string_value", "5"),
    (str, 0, "string_value", "0"),
])
def test_default_is_kept_and_parameter_is_optional(build_parameter_type_spec, ptype, default, kind, expected):
    spec = build_parameter_type_spec(data_types.RuntimeParameter(name="p", ptype=ptype, default=default))

    assert spec.default_value.WhichOneof("kind") == kind
    assert getattr(spec.default_value, kind) == expected
    assert spec.is_optional is True


@pytest.mark.parametrize("value", [
    data_types.RuntimeParameter(name="p", ptype=str),
    data_types.RuntimeParameter(name="p", ptype=int, default=""),
    "constant",
    5,
], ids=["no-default", "falsy-default-of-another-type", "string-constant", "int-constant"])
def test_parameter_without_a_usable_default_stays_required(build_parameter_type_spec, value):
    spec = build_parameter_type_spec(value)

    assert not spec.HasField("default_value")
    assert spec.is_optional is False


def test_original_builder_is_restored():
    original = compiler_utils.build_parameter_type_spec

    with runtime_parameter_defaults():
        assert compiler_utils.build_parameter_type_spec is not original

    assert compiler_utils.build_parameter_type_spec is original


def test_original_builder_is_restored_after_an_error():
    original = compiler_utils.build_parameter_type_spec

    with pytest.raises(RuntimeError), runtime_parameter_defaults():
        raise RuntimeError

    assert compiler_utils.build_parameter_type_spec is original
