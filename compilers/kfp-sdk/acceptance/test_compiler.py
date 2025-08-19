import os
import sys
import yaml
import json

import pytest
from click.testing import CliRunner
from compiler import compiler
from tempfile import TemporaryDirectory


runner = CliRunner()


@pytest.fixture(scope="session", autouse=True)
def setup():
    sys.path.append(
        os.path.join(
            os.path.dirname(__file__),
            "..",
            "..",
            "..",
            "docs-gen",
            "includes",
            "master",
            "kfpsdk-quickstart",
        )
    )


def test_compiler__compile(tmp_path):
    with TemporaryDirectory() as tmp_dir:
        output_file = os.path.join(tmp_dir, "pipeline.yaml")

        pipeline_config = "acceptance/pipeline.yaml"

        result = runner.invoke(
            compiler.compile,
            [
                "--pipeline_config",
                pipeline_config,
                "--output_file",
                str(output_file),
            ],
        )

        assert result.exit_code == 0
        assert os.stat(output_file).st_size != 0

        with open(output_file, "r") as f:
            pipeline = yaml.safe_load(f.read())

        assert "labels" in pipeline
        assert pipeline["pipelineSpec"]["schemaVersion"] == "2.1.0"
        assert pipeline["displayName"] == "test"


def test_compiler__compile_json_output(tmp_path):
    """Test that the compiler can output JSON format when outputfile extension is .json"""
    with TemporaryDirectory() as tmp_dir:
        output_file = os.path.join(tmp_dir, "pipeline.json")

        pipeline_config = "acceptance/pipeline.yaml"

        result = runner.invoke(
            compiler.compile,
            [
                "--pipeline_config",
                pipeline_config,
                "--output_file",
                str(output_file),
            ],
        )

        assert result.exit_code == 0
        assert os.stat(output_file).st_size != 0

        with open(output_file, "r") as f:
            content = f.read()
            pipeline = json.loads(content)

        assert "labels" in pipeline
        assert "runtimeConfig" in pipeline
        assert pipeline["pipelineSpec"]["schemaVersion"] == "2.1.0"
        assert pipeline["displayName"] == "test"


def test_compiler__unsupported_extension(tmp_path):
    """Test that the compiler raises an error for unsupported file extensions"""
    with TemporaryDirectory() as tmp_dir:
        output_file = os.path.join(tmp_dir, "pipeline.txt")

        pipeline_config = "acceptance/pipeline.yaml"

        result = runner.invoke(
            compiler.compile,
            [
                "--pipeline_config",
                pipeline_config,
                "--output_file",
                str(output_file),
            ],
        )

        assert result.exit_code != 0
        assert "Unsupported output file format" in str(result.exception)
        assert "Expected .json, .yaml, or .yml" in str(result.exception)
