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
        provider_config = "acceptance/provider.yaml"

        result = runner.invoke(
            compiler.compile,
            [
                "--pipeline_config",
                pipeline_config,
                "--provider_config",
                provider_config,
                "--output_file",
                str(output_file),
            ],
        )

        assert result.exit_code == 0
        assert os.stat(output_file).st_size != 0

        with open(output_file, "r") as f:
            pipeline = yaml.safe_load(f.read())

        assert pipeline["schemaVersion"] == "2.1.0"
        assert pipeline["defaultPipelineRoot"] == "gs://bucket/test"
