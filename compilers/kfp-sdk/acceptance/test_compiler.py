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


@pytest.mark.parametrize(
    "file_extension,loader",
    [
        ("yaml", yaml.safe_load),
        ("yml", yaml.safe_load),
        ("json", json.loads),
    ],
)
def test_compiler__compile_with_different_formats(file_extension, loader):
    """Test that the compiler can output different file formats"""
    with TemporaryDirectory() as tmp_dir:
        output_file = os.path.join(tmp_dir, f"pipeline.{file_extension}")
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
            pipeline = loader(f.read())

        assert "labels" in pipeline
        assert pipeline["pipelineSpec"]["schemaVersion"] == "2.1.0"
        assert pipeline["displayName"] == "test"


def test_compiler__unsupported_extension():
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
