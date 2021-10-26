from click.testing import CliRunner
from unittest.mock import Mock
from mlmd_cli.cli import pushed_model_artifact
from mlmd_cli.mlmd_client import MlmdClient
import unittest
from ml_metadata.proto import metadata_store_service_pb2, metadata_store_pb2

class TestJob(unittest.TestCase):
    runner = CliRunner()
    workflow_name = 'aWorkflow'
    pipeline_name = 'aPipeline'
    context_id    = 12345
    execution_id  = 23456
    artifact_id   = 34567
    artifact_location = 'some://where'

    def test_get_pushed_model(self):
        grpc_client = grpc_client=Mock()

        grpc_client.GetContextByTypeAndName.return_value = self.context_response(context_id=self.context_id)
        grpc_client.GetExecutionsByContext.return_value = self.executions_response(run_id=self.workflow_name, execution_id=self.execution_id)
        grpc_client.GetEventsByExecutionIDs.return_value = self.events_response(artifact_id=self.artifact_id)
        grpc_client.GetArtifactsByID.return_value = self.artifacts_response(pushed_destination=self.artifact_location)

        result = self.runner.invoke(pushed_model_artifact, [
                'get',
                '--workflow-name', self.workflow_name,
                '--pipeline-name', self.pipeline_name,      
            ], obj={'client': MlmdClient(grpc_client),})

        # self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.output, f"{self.artifact_location}\n")

    def executions_response(self, run_id, execution_id):
        execution = metadata_store_pb2.Execution()
        execution.properties['run_id'].string_value = run_id
        execution.properties['component_id'].string_value = 'Pusher'
        execution.id = execution_id
        executions_response = metadata_store_service_pb2.GetExecutionsByContextResponse()
        executions_response.executions.extend([execution])
        return executions_response
    
    def context_response(self, context_id):
        context_response = metadata_store_service_pb2.GetContextByTypeAndNameResponse()
        context_response.context.id = context_id

        return context_response

    def events_response(self, artifact_id):
        event = metadata_store_pb2.Event()
        step = metadata_store_pb2.Event.Path.Step()
        step.key = 'pushed_model'
        event.path.steps.extend([step])
        event.artifact_id = artifact_id
        events_response = metadata_store_service_pb2.GetEventsByExecutionIDsResponse()
        events_response.events.extend([event])

        return events_response

    def artifacts_response(self, pushed_destination):
        artifact = metadata_store_pb2.Artifact()
        artifact.custom_properties['pushed_destination'].string_value = pushed_destination
        artifacts_response = metadata_store_service_pb2.GetArtifactsByIDResponse()
        artifacts_response.artifacts.extend([artifact])

        return artifacts_response
