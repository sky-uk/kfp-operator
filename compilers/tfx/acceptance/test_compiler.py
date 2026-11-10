from compiler import compiler
from click.testing import CliRunner
import yaml
import json
import os
import sys
from tempfile import TemporaryDirectory
import pytest

runner = CliRunner()
config_file_path = 'acceptance/pipeline_conf.yaml'


@pytest.fixture(scope="session", autouse=True)
def setup():
    sys.path.append(os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs-gen", "includes", "master", "quickstart", "penguin_pipeline"))


def test_cli():
    with TemporaryDirectory() as tmp_dir:
        output_file_path = os.path.join(tmp_dir, 'pipeline.json')

        result = runner.invoke(compiler.compile, ['--pipeline_config', config_file_path, '--output_file', output_file_path])

        assert result.exit_code == 0
        assert os.stat(output_file_path).st_size != 0

        f = open(output_file_path, "r")
        pipeline = json.load(f)
        assert pipeline['pipelineSpec']['schemaVersion'] == '2.1.0'
        assert pipeline['pipelineSpec']['pipelineInfo']['name'] == "namespace-test"
        assert pipeline['pipelineSpec']['root']['inputDefinitions']['parameters']['push_destination'] == {
            'parameterType': 'STRING',
        }


def test_runtime_parameter_defaults_are_kept():
    with TemporaryDirectory() as tmp_dir:
        output_file_path = os.path.join(tmp_dir, 'pipeline.yaml')

        result = runner.invoke(compiler.compile, ['--pipeline_config', 'acceptance/runtime_parameter_pipeline_conf.yaml', '--output_file', output_file_path])

        assert result.exit_code == 0

        with open(output_file_path) as f:
            pipeline_spec = yaml.safe_load(f)
        assert pipeline_spec['schemaVersion'] == '2.1.0'
        assert pipeline_spec['root']['inputDefinitions']['parameters']['input_base'] == {
            'parameterType': 'STRING',
            'defaultValue': '/data',
            'isOptional': True,
        }


def test_failure():
    result = runner.invoke(compiler.compile, ['--pipeline_config', ''])
    assert result.exit_code != 0
    assert os.path.isfile('not_existing.yaml') == False
