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

provider_config_file_path = 'acceptance/provider_conf.yaml'


@pytest.fixture(scope="session", autouse=True)
def setup():
    sys.path.append(os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs-gen", "includes", "master", "quickstart", "penguin_pipeline"))


def test_cli():
    with TemporaryDirectory() as tmp_dir:
        output_file_path = os.path.join(tmp_dir, 'pipeline.yaml')

        result = runner.invoke(compiler.compile, ['--pipeline_config', config_file_path, '--provider_config', provider_config_file_path, '--output_file', output_file_path])

        assert result.exit_code == 0
        assert os.stat(output_file_path).st_size != 0

        f = open(output_file_path, "r")
        pipeline = yaml.safe_load(f.read())
        assert pipeline['pipelineSpec']['schemaVersion'] == '2.0.0'
        assert pipeline['pipelineSpec']['pipelineInfo']['name'] == "namespace-test"


def test_failure():
    result = runner.invoke(compiler.compile, ['--pipeline_config', ''])
    assert result.exit_code != 0
    assert os.path.isfile('not_existing.yaml') == False
