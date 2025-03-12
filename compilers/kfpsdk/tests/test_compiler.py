import os
import sys
import uuid

from compiler import compiler
from kfp import dsl


def pipeline_fn():
    return True


@dsl.component(base_image="python:3.10")
def component():
    pass


@dsl.pipeline(
    name='Test pipeline',
    description='Empty pipeline.')
def pipeline_fn2():
    component()


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


def test_compiler__compile(tmp_path):
    pipeline_config = temp_yaml_file(tmp_path)
    pipeline_config.write_text('framework:\n  parameters:\n    pipeline: test_compiler.pipeline_fn2\n')
    output_file = temp_yaml_file(tmp_path)

    compiler._compile(str(pipeline_config), str(output_file))

    assert output_file.exists()
    assert output_file.stat().st_size > 0


def temp_yaml_file(tmp_path):
    return tmp_path / f"{uuid.uuid4()}.yaml"
