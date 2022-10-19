from kfp_compiler import compiler
from click.testing import CliRunner
import yaml
import json
import os
import sys
from tempfile import TemporaryDirectory

runner = CliRunner()
config_file_path = 'acceptance/pipeline_conf.yaml'

def test_cli_v1():
    with TemporaryDirectory() as tmp_dir:
        output_file_path = os.path.join(tmp_dir, 'pipeline.yaml')

        sys.path.append(os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs-gen", "includes", "quickstart", "penguin_pipeline"))
        result = runner.invoke(compiler.compile, ['--pipeline_config', config_file_path, '--output_file', output_file_path, "--execution_mode", "v1"])

        assert result.exit_code == 0
        assert os.stat(output_file_path).st_size != 0

        f = open(output_file_path, "r")
        workflow = yaml.load(f.read())
        assert workflow['kind'] == 'Workflow'

def test_cli_v2():
    with TemporaryDirectory() as tmp_dir:
        output_file_path = os.path.join(tmp_dir, 'pipeline.json')
        sys.path.append(os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs-gen", "includes", "quickstart", "penguin_pipeline"))
        result = runner.invoke(compiler.compile, ['--pipeline_config', config_file_path, '--output_file', output_file_path, "--execution_mode", "v2"])

        assert result.exit_code == 0
        assert os.stat(output_file_path).st_size != 0

        f = open(output_file_path, "r")
        pipeline = json.loads(f.read())
        assert pipeline['pipelineSpec']['schemaVersion'] == '2.0.0'


def test_failure():
    result = runner.invoke(compiler.compile, ['--pipeline_config', ''])
    assert result.exit_code != 0
    assert os.path.isfile('not_existing.yaml') == False
