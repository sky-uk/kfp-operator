import os
import sys
import uuid

import pytest
from click.testing import CliRunner
from compiler import compiler

runner = CliRunner()


@pytest.fixture(scope="session", autouse=True)
def setup():
    sys.path.append(
        os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs-gen", "includes", "master", "kfpsdk-quickstart",
                     "getting_started"))


def test_compiler__compile(tmp_path):
    pipeline_config = 'acceptance/pipeline.yaml'
    provider_config = 'acceptance/provider.yaml'
    output_file = temp_yaml_file(tmp_path)

    result = runner.invoke(compiler.compile,
                           ['--pipeline_config', pipeline_config, '--provider_config', provider_config, '--output_file',
                            str(output_file)])

    assert result.exit_code == 0
    assert os.stat(output_file).st_size != 0


def temp_yaml_file(tmp_path):
    return tmp_path / f"{uuid.uuid4()}.yaml"
