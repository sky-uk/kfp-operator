import os
import sys

import pytest
from compiler import compiler


def pipeline_fn():
    return True


def test_compiler_load_fn():
    pipeline_config_contents = {
        'framework': {
            'parameters': {
                'pipeline': 'test_compiler.pipeline_fn'
            }
        }
    }
    sys.path.append(os.path.dirname(__file__))
    result = compiler.load_fn(pipeline_config_contents)()
    assert result

def test_compiler_missing_pipeline_parameter():
    pipeline_config_contents = {
        'framework': {
            'parameters': {}
        }
    }
    sys.path.append(os.path.dirname(__file__))
    with pytest.raises(KeyError) as error:
        compiler.load_fn(pipeline_config_contents)()

    assert str(error.value) == "'Missing required framework parameter: [pipeline].'"

def test_compiler_invalid_pipeline_format():
    pipeline_config_contents = {
        'framework': {
            'parameters': {
                'pipeline': 'function'
            }
        }
    }
    sys.path.append(os.path.dirname(__file__))
    with pytest.raises(ValueError) as error:
        compiler.load_fn(pipeline_config_contents)()

    assert str(error.value) == "Invalid pipeline format: [function]. Expected format: 'module_path.function_name'."
