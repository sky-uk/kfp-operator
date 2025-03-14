import importlib
import importlib.util

import click
import yaml
from kfp import compiler


@click.command()
@click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
@click.option('--provider_config', help='Provider configuration in yaml format', required=False)
@click.option('--output_file', help='Output file path', required=True)
def compile(pipeline_config: str, provider_config: str, output_file: str):
    _compile(pipeline_config, output_file)


def _compile(pipeline_config: str, output_file: str):
    """Compiles KFP SDK pipeline into a Kubeflow Pipelines pipeline definition"""
    with open(pipeline_config, "r") as pipeline_stream:
        pipeline_config_contents = yaml.safe_load(pipeline_stream)
        click.secho(f'Compiling with pipeline: {pipeline_config_contents}', fg='green')

        pipeline_fn = load_fn(pipeline_config_contents)

        compiler.Compiler().compile(pipeline_fn, package_path=output_file)
        click.secho(f'{output_file} compiled', fg='green')


def load_fn(pipeline_config_contents: dict):
    framework_parameters = pipeline_config_contents['framework']['parameters']
    if 'pipeline' not in framework_parameters:
        raise KeyError('Missing required framework parameter: [pipeline].')
    pipeline = framework_parameters['pipeline']

    if "." not in pipeline:
        raise ValueError(f"Invalid pipeline format: [{pipeline}]. Expected format: 'module_path.function_name'.")

    (module_name, fn_name) = pipeline.rsplit('.', 1)
    module = importlib.import_module(module_name)

    loaded_module = dir(module)
    click.secho(f'Loaded module: {loaded_module}', fg='green')

    fn = getattr(module, fn_name)

    return fn
