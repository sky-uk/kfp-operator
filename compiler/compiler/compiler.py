import click
import yaml
from tfx.orchestration.kubeflow import kubeflow_dag_runner

import sys
import importlib
from tfx.orchestration import pipeline
from compiler import component_expander, module_loader

@click.command()
@click.option('--pipeline_config', help='Pipeline configuration file path in yaml format', required=True)
@click.option('--output_file', help='Output file path', required=True)
def compile(pipeline_config, output_file):
    """Compiles TFX components into a Kubeflow Pipelines pipeline definition"""
    with open(pipeline_config, 'r') as stream:
        config = yaml.safe_load(stream)

        click.secho(f'Compiling with config: {config}', fg='green')
        spec = config['spec']
        
        components = module_loader.load_fn(spec['tfx_components'], spec['env'])()
        expanded_components = component_expander.expand_components_with_pusher(components, config['serving_dir'])

        metadata_config = kubeflow_dag_runner.get_default_kubeflow_metadata_config()

        runner_config = kubeflow_dag_runner.KubeflowDagRunnerConfig(
            kubeflow_metadata_config=metadata_config, tfx_image=spec['image']
        )

        kubeflow_dag_runner.KubeflowDagRunner(
            config=runner_config, output_filename=output_file
        ).run(
            pipeline.Pipeline(
                pipeline_name=config['name'],
                pipeline_root=config['pipeline_root'],
                components=expanded_components,
                enable_cache=False,
                metadata_connection_config=None,
                beam_pipeline_args=dict_to_cli_args(config['beam_args'])
            )
        )

        click.secho(f'{output_file} written', fg='green')

def dict_to_cli_args(beam_args):
    return [f'--{k}={v}' for k,v in beam_args.items()]
        

if __name__ == '__main__':
    compile()