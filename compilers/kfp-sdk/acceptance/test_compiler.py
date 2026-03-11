import os
import sys
import yaml

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

        assert pipeline["schemaVersion"] == "2.1.0"

        # Assert that all components use the image from pipeline config
        executors = pipeline["deploymentSpec"]["executors"]
        expected_image = (
            "python:3.9"  # This should match what's in acceptance/pipeline.yaml
        )

        for executor_name, executor_spec in executors.items():
            actual_image = executor_spec["container"]["image"]
            assert (
                actual_image == expected_image
            ), f"Executor '{executor_name}' has image '{actual_image}', expected '{expected_image}'"
