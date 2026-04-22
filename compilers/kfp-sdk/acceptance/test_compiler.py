import os
import sys
from tempfile import TemporaryDirectory

import pytest
import yaml
from click.testing import CliRunner

from compiler import compiler

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

        with open(output_file) as f:
            pipeline = yaml.safe_load(f.read())

        assert pipeline["schemaVersion"] == "2.1.0"

        executors = pipeline["deploymentSpec"]["executors"]
        expected_image = "foo:1.2.3"

        for executor_name, executor_spec in executors.items():
            actual_image = executor_spec["container"]["image"]
            assert actual_image == expected_image, (
                f"Executor '{executor_name}' has image"
                f" '{actual_image}', expected '{expected_image}'"
            )
