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


@pytest.mark.parametrize("use_pipeline_spec_2_1,expected_schema_version", [
    (True, '2.1.0'),
    (False, '2.0.0'),
])
def test_cli_yaml_output(use_pipeline_spec_2_1, expected_schema_version):
    with TemporaryDirectory() as tmp_dir:
        with open(config_file_path) as config_stream:
            config = yaml.safe_load(config_stream)
        config['framework']['parameters']['usePipelineSpec2_1'] = use_pipeline_spec_2_1
        config_path = os.path.join(tmp_dir, 'pipeline_conf.yaml')
        with open(config_path, 'w') as config_stream:
            yaml.safe_dump(config, config_stream)

        output_file_path = os.path.join(tmp_dir, 'pipeline.yaml')

        result = runner.invoke(compiler.compile, ['--pipeline_config', config_path, '--output_file', output_file_path])

        assert result.exit_code == 0, result.output
        with open(output_file_path) as output_stream:
            pipeline = yaml.safe_load(output_stream)
        assert pipeline['pipelineSpec']['schemaVersion'] == expected_schema_version
        assert pipeline['pipelineSpec']['pipelineInfo']['name'] == "namespace-test"


def test_failure():
    result = runner.invoke(compiler.compile, ['--pipeline_config', ''])
    assert result.exit_code != 0
    assert os.path.isfile('not_existing.yaml') == False
