import importlib
import importlib.util
import os

import click
import yaml
from kfp import compiler
from typing import Callable


@click.command()
@click.option(
    "--pipeline_config", help="Pipeline configuration in yaml format", required=True
)
@click.option(
    "--provider_config", help="Provider configuration in yaml format", required=True
)
@click.option("--output_file", help="Output file path", required=True)
def compile(pipeline_config: str, provider_config: str, output_file: str):
    """Compiles KFP SDK pipeline into a Kubeflow Pipelines pipeline definition"""
    with open(pipeline_config, "r") as pipeline_stream, open(
        provider_config, "r"
    ) as provider_stream:
        pipeline_config_contents = yaml.safe_load(pipeline_stream)
        pipeline_name = get_pipeline_name(pipeline_config_contents)
        sanitised_pipeline_name = sanitise_namespaced_pipeline_name(pipeline_name)
        pipeline_environment = pipeline_config_contents.get("env", [])
        click.secho(
            f"Compiling {sanitised_pipeline_name} pipeline with root {pipeline_config_contents}",
            fg="green",
        )

        pipeline_fn = load_fn(pipeline_config_contents, pipeline_environment)

        compiler.Compiler().compile(
            pipeline_fn, pipeline_name=sanitised_pipeline_name, package_path=output_file
        )
        click.secho(f"{output_file} compiled", fg="green")


def get_pipeline_name(pipeline_config_contents: dict) -> str:
    if "name" in pipeline_config_contents:
        return pipeline_config_contents["name"]
    else:
        raise KeyError("Missing required pipeline name in pipeline configuration.")


def load_fn(pipeline_config_contents: dict, environment: list) -> Callable:
    framework_parameters = pipeline_config_contents["framework"]["parameters"]
    if "pipeline" not in framework_parameters:
        raise KeyError("Missing required framework parameter: [pipeline].")
    pipeline = framework_parameters["pipeline"]

    if "." not in pipeline:
        raise ValueError(
            f"Invalid pipeline format: [{pipeline}]. Expected format: 'module_path.function_name'."
        )

    for env in environment:
        os.environ[env["name"]] = env["value"]

    (module_name, fn_name) = pipeline.rsplit(".", 1)
    module = importlib.import_module(module_name)

    loaded_module = dir(module)
    click.secho(f"Loaded module: {loaded_module}", fg="green")

    fn = getattr(module, fn_name)

    return fn


def sanitise_namespaced_pipeline_name(namespaced_name: str) -> str:
    return namespaced_name.replace("/", "-")
