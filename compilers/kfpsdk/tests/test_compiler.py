import os
import sys

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
