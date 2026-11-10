def apply_runtime_parameter_default(parameter_spec, value):
    """Copy a TFX RuntimeParameter default onto a KFP ParameterSpec.

    TFX records the default on RuntimeParameter and then drops it when building
    the spec. A missing default makes the parameter required.
    """
    default = getattr(value, "default", None)
    if default is None:
        return parameter_spec

    if isinstance(default, bool):
        parameter_spec.default_value.bool_value = default
    elif isinstance(default, int):
        parameter_spec.default_value.number_value = default
    elif isinstance(default, float):
        parameter_spec.default_value.number_value = default
    elif isinstance(default, str):
        parameter_spec.default_value.string_value = default
    else:
        return parameter_spec

    parameter_spec.is_optional = True
    return parameter_spec


def install_runtime_parameter_defaults():
    """Make TFX's parameter spec builders keep RuntimeParameter defaults."""
    import tfx.orchestration.kubeflow.v2.compiler_utils as compiler_utils

    for name in ("build_parameter_type_spec", "build_parameter_type_spec_legacy"):
        original = getattr(compiler_utils, name, None)
        if original is None or getattr(original, "_applies_runtime_parameter_default", False):
            continue

        def build(value, build_spec=original):
            return apply_runtime_parameter_default(build_spec(value), value)

        build._applies_runtime_parameter_default = True
        setattr(compiler_utils, name, build)
