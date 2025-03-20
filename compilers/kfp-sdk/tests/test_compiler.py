import os
import sys

import pytest
from compiler import compiler


def pipeline_fn():
    return True


def test_compiler_load_fn():
    pipeline_config_contents = {
        "framework": {"parameters": {"pipeline": "test_compiler.pipeline_fn"}}
    }
    sys.path.append(os.path.dirname(__file__))
    result = compiler.load_fn(pipeline_config_contents, environment=[])()
    assert result


def pipeline_fn_env():
    return os.environ


def test_compiler_load_fn_env():

    environment = [
        {"name": "a", "value": "aVal"},
    ]

    pipeline_config_contents = {
        "framework": {"parameters": {"pipeline": "test_compiler.pipeline_fn_env"}},
        "env": environment,
    }

    sys.path.append(os.path.dirname(__file__))
    result = compiler.load_fn(pipeline_config_contents, environment)()

    assert result["a"] == "aVal"


def test_compiler_missing_pipeline_parameter():
    pipeline_config_contents = {"framework": {"parameters": {}}}
    sys.path.append(os.path.dirname(__file__))
    with pytest.raises(KeyError) as error:
        compiler.load_fn(pipeline_config_contents, environment=[])

    assert str(error.value) == "'Missing required framework parameter: [pipeline].'"


def test_compiler_invalid_pipeline_format():
    pipeline_config_contents = {"framework": {"parameters": {"pipeline": "function"}}}
    sys.path.append(os.path.dirname(__file__))
    with pytest.raises(ValueError) as error:
        compiler.load_fn(pipeline_config_contents, environment=[])()

    assert (
        str(error.value)
        == "Invalid pipeline format: [function]. Expected format: 'module_path.function_name'."
    )


def test_sanitise_namespaced_pipeline_name():
    tests = [
        ("namespace/name", "namespace-name"),
        ("test", "test"),
        ("/", "-"),
        ("a/b/c/d/e/f/g/h/", "a-b-c-d-e-f-g-h-"),
    ]

    for input, output in tests:
        sanitised = compiler.sanitise_namespaced_pipeline_name(input)
        assert sanitised == output, f"Expected {output}, got {sanitised}"


def test_compiler_get_pipeline_root():
    pipeline_config_contents = {"name": "pipeline"}
    provider_config_contents = {"pipelineRootStorage": "storage"}
    assert (
        compiler.get_pipeline_root(pipeline_config_contents, provider_config_contents)
        == "storage/pipeline"
    )
