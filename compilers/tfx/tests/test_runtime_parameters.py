from compiler.runtime_parameters import apply_runtime_parameter_default


class _Value:
    def __init__(self):
        self.string_value = ""
        self.number_value = 0
        self.bool_value = False


class _Spec:
    def __init__(self):
        self.default_value = _Value()
        self.is_optional = False


class _Parameter:
    def __init__(self, default):
        self.default = default


def test_string_default_is_copied_and_optional():
    spec = apply_runtime_parameter_default(_Spec(), _Parameter("us"))

    assert spec.default_value.string_value == "us"
    assert spec.is_optional is True


def test_empty_string_and_zero_are_kept():
    empty = apply_runtime_parameter_default(_Spec(), _Parameter(""))
    zero = apply_runtime_parameter_default(_Spec(), _Parameter(0))

    assert empty.default_value.string_value == ""
    assert empty.is_optional is True
    assert zero.default_value.number_value == 0
    assert zero.is_optional is True


def test_missing_default_stays_required():
    spec = apply_runtime_parameter_default(_Spec(), _Parameter(None))

    assert spec.default_value.string_value == ""
    assert spec.is_optional is False


def test_value_without_a_default_attribute_is_unchanged():
    spec = apply_runtime_parameter_default(_Spec(), "constant")

    assert spec.is_optional is False
