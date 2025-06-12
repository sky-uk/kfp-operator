import click
import yaml
import sys
import importlib
import importlib.util
import os
from tfx.v1.orchestration import experimental as kubeflow_dag_runner
from tfx.orchestration import pipeline


@click.command()
@click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
@click.option('--provider_config', help='Provider configuration in yaml format', required=True)
@click.option('--output_file', help='Output file path', required=True)
def compile(pipeline_config: str, provider_config: str, output_file: str):
    """Compiles TFX components into a Kubeflow Pipelines pipeline definition"""
    with open(pipeline_config, "r") as pipeline_stream, open(provider_config, "r") as provider_stream:
        pipeline_config_contents = yaml.safe_load(pipeline_stream)
        provider_config_contents = yaml.safe_load(provider_stream)

        click.secho(f'Compiling with pipeline: {pipeline_config_contents} and provider {provider_config_contents} ',
                    fg='green')

        pipeline_root, temp_location = pipeline_paths_for_config(pipeline_config_contents, provider_config_contents)

        framework_parameters = pipeline_config_contents['framework']['parameters']
        pipeline_beam_args = framework_parameters.get('beamArgs', [])
        beam_cli_args = name_values_to_cli_args(pipeline_beam_args)
        beam_cli_args.append(f"--temp_location={temp_location}")

        components = load_fn(framework_parameters.get('components', ""), pipeline_config_contents.get('env', []))()

        compile_v2(pipeline_config_contents, output_file).run(
            pipeline.Pipeline(
                pipeline_name=sanitise_namespaced_pipeline_name(pipeline_config_contents['name']),
                pipeline_root=pipeline_root,
                components=components,
                enable_cache=False,
                metadata_connection_config=None,
                beam_pipeline_args=beam_cli_args
            )
        )

        click.secho(f'{output_file} written', fg='green')


def load_fn(tfx_components: str, env: list):
    for name_value in env:
        os.environ[name_value['name']] = name_value['value']

    if "." not in tfx_components:
        raise ValueError(
            f"Invalid tfxComponents format: [{tfx_components}]. Expected format: 'module_path.function_name'."
        )

    (module_name, fn_name) = tfx_components.rsplit('.', 1)
    module = importlib.import_module(module_name)
    fn = getattr(module, fn_name)

    return fn


def sanitise_namespaced_pipeline_name(namespaced_name: str) -> str:
    return namespaced_name.replace("/", "-")


def compile_v2(config: dict, output_filename: str):
    runner_config = kubeflow_dag_runner.KubeflowV2DagRunnerConfig(
        display_name=config['name'],
        default_image=config['image']
    )

    return kubeflow_dag_runner.KubeflowV2DagRunner(
        config=runner_config,
        output_filename=output_filename
    )


def pipeline_paths_for_config(pipeline_config: dict, provider_config: dict):
    pipeline_root = provider_config['pipelineRootStorage'] + '/' + pipeline_config['name']
    return pipeline_root, pipeline_root + "/tmp"


def name_values_to_cli_args(name_values: list):
    cli_args = []

    for name_value in name_values:
        cli_args.append(f'--{name_value["name"]}={name_value["value"]}')

    return cli_args


def main():
    compile()
