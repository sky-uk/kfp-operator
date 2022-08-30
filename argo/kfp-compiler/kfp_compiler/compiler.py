import click
import yaml
import sys
import importlib
import importlib.util
import os
from tfx import v1 as tfx
from tfx.v1.dsl import Pipeline
from tfx.components import Pusher, Trainer
from tfx.v1.proto import PushDestination

def expand_components_with_pusher(tfx_components, serving_model_directory):
    if not any(isinstance(component, Pusher) for component in tfx_components):

        click.secho(
            "Could not find Pusher component. Trying to expand. Expanding", fg="green"
        )

        trainers = (
            comp
            for comp in tfx_components
            if isinstance(comp, Trainer) and "model" in comp.outputs
        )
        trainer = next(trainers, None)

        if next(trainers, None):
            click.secho("Found more than one Trainer component. aborting", fg="red")
            sys.exit(1)

        if trainer:
            model = trainer.outputs["model"]
        else:
            return tfx_components

        evaluators = (comp for comp in tfx_components if "blessing" in comp.outputs)
        evaluator = next(evaluators, None)

        if next(evaluators, None):
            click.secho("Found more than one Evaluator component. aborting", fg="red")
            sys.exit(1)

        blessing = evaluator.outputs["blessing"] if evaluator else None

        pusher = Pusher(
            model=model,
            model_blessing=blessing,
            push_destination=PushDestination(
                filesystem=PushDestination.Filesystem(
                    base_directory=serving_model_directory
                )
            ),
        )

        return tfx_components + [pusher]
    else:
        return tfx_components


def load_fn(tfx_components, env={}):
    for key, value in env.items():
        os.environ[key] = value

    (module_name, fn_name) = tfx_components.rsplit('.', 1)
    module = importlib.import_module(module_name)
    fn = getattr(module, fn_name)

    return fn


@click.command()
@click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
@click.option('--output_file', help='Output file path', required=True)
def compile(pipeline_config, output_file):
    """Compiles TFX components into a Kubeflow Pipelines pipeline definition"""
    config = yaml.safe_load(pipeline_config)

    click.secho(f'Compiling with config: {config}', fg='green')

    components = load_fn(config['tfxComponents'], config.get('env', {}))()
    expanded_components = expand_components_with_pusher(components, config['servingLocation'])

    metadata_config = tfx.orchestration.experimental.get_default_kubeflow_metadata_config()

    runner_config = tfx.orchestration.experimental.KubeflowDagRunnerConfig(
        kubeflow_metadata_config=metadata_config, tfx_image=config['image']
    )

    tfx.orchestration.experimental.KubeflowDagRunner(
        config=runner_config, output_filename=output_file
    ).run(
        Pipeline(
            pipeline_name=config['name'],
            pipeline_root=config['rootLocation'],
            components=expanded_components,
            enable_cache=False,
            metadata_connection_config=None,
            beam_pipeline_args=dict_to_cli_args(config.get('beamArgs', {}))
        )
    )

    click.secho(f'{output_file} written', fg='green')


def dict_to_cli_args(beam_args):
    beam_cli_args = []
    for k, v in beam_args.items():
        for vv in v:
            beam_cli_args.append(f'--{k}={vv}')

    return beam_cli_args


def main():
    compile()
