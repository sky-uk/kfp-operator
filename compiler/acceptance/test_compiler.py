from compiler import compiler
from click.testing import CliRunner
import yaml
import os
import sys
import tempfile
dirpath = tempfile.mkdtemp()

runner = CliRunner()
config_file_path = 'acceptance/pipeline_conf.yaml'
output_file_path = dirpath+'/pipeline.yaml'

def test_cli():
    sys.path.append(os.path.dirname(__file__))
    with open(config_file_path, 'r') as f:        
        result = runner.invoke(compiler.compile, ['--pipeline_config', f.read(), '--output_file', output_file_path])

        assert result.exit_code == 0
        assert os.stat(output_file_path).st_size != 0


def test_failure():
    result = runner.invoke(compiler.compile, ['--pipeline_config', ''])
    assert result.exit_code != 0
    assert os.path.isfile('not_existing.yaml') == False