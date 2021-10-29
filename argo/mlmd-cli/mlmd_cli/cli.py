import click

from grpc import insecure_channel
from ml_metadata.proto import metadata_store_service_pb2_grpc, metadata_store_service_pb2_grpc
from .mlmd_client import MlmdClient

@click.group()
@click.option('--endpoint', help='Endpoint of the ML Meta data gRPC server.', required=True)
@click.pass_context
def cli(ctx, endpoint):
  channel = insecure_channel(endpoint)
  ctx.obj['client'] = MlmdClient(grpc_client=metadata_store_service_pb2_grpc.MetadataStoreServiceStub(channel))

@click.group()
def pushed_model_artifact():
  """Manage pushed model artifacts."""
  pass

@pushed_model_artifact.command()
@click.option('--workflow-name', help='Name of the Kubeflow Argo workflow.', required=True)
@click.option('--pipeline-name', help='Name of the Kubeflow Pipeline.', required=True)
@click.pass_context
def get(ctx, workflow_name, pipeline_name):
  client = ctx.obj['client']

  pipeline_id = client.get_context_id(context_type='pipeline', context_name=pipeline_name)
  execution_id = client.get_execution_id(context_id=pipeline_id, run_id=workflow_name, component_id='Pusher')
  artifact_id = client.get_artifact_id(execution_id=execution_id, step_name='pushed_model')
  destination = client.get_artifact_property(artifact_id, property_name='pushed_destination')

  click.echo(destination)

def main():
    cli.add_command(pushed_model_artifact)
    cli(obj={})
