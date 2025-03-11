import click
import yaml
import importlib
import importlib.util
from kfp import compiler

@click.command()
@click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
@click.option('--provider_config', help='Provider configuration in yaml format', required=True)
@click.option('--output_file', help='Output file path', required=True)
def compile(pipeline_config: str, provider_config: str, output_file: str):
    """Compiles KFP SDK pipeline into a Kubeflow Pipelines pipeline definition"""
    with open(pipeline_config, "r") as pipeline_stream, open(provider_config, "r") as provider_stream:
        pipeline_config_contents = yaml.safe_load(pipeline_stream)
        provider_config_contents = yaml.safe_load(provider_stream)

        click.secho(f'Compiling with pipeline: {pipeline_config_contents} and provider {provider_config_contents} ',
                    fg='green')
        
        pipeline_root, serving_model_directory, temp_location = pipeline_paths_for_config(pipeline_config_contents, provider_config_contents)

        pipeline_fn = load_fn(pipeline_config_contents)
        
        # TODO: pusher?

        compiler.Compiler().compile(pipeline_fn, package_path=output_file)

        click.secho(f'{output_file} compiled', fg='green')

def load_fn(pipeline_config_contents: dict):
    framework = pipeline_config_contents['framework']
    framework_parameters = framework['parameters']
    pipeline = framework_parameters['pipeline']

    (module_name, fn_name) = pipeline.rsplit('.', 1)
    module = importlib.import_module(module_name)
    
    loaded_module = dir(module)
    click.secho(f'Loaded module: {loaded_module}', fg='green')

    fn = getattr(module, fn_name)

    return fn

def pipeline_paths_for_config(pipeline_config: dict, provider_config: dict):
    pipeline_root = provider_config['pipelineRootStorage'] + '/' + pipeline_config['name']
    return pipeline_root, pipeline_root + "/serving", pipeline_root + "/tmp"
