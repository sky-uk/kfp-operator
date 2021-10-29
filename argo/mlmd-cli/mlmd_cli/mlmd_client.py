import click
from ml_metadata.proto import metadata_store_service_pb2

class MlmdClient():
    def __init__(self, grpc_client):
        self.grpc_client = grpc_client

    def get_context_id(self, context_type, context_name):
        request = metadata_store_service_pb2.GetContextByTypeAndNameRequest()
        request.type_name = context_type
        request.context_name = context_name
        response = self.grpc_client.GetContextByTypeAndName(request)
        if response.context.id != 0:
            return response.context.id
        else:
            raise RuntimeError(f"No context for {context_type} {context_name} could be found")

    def get_execution_id(self, context_id, run_id, component_id):
        request = metadata_store_service_pb2.GetExecutionsByContextRequest()
        request.context_id = context_id
        response = self.grpc_client.GetExecutionsByContext(request)
        return first_or_raise(
            [e.id for e in response.executions if e.properties['run_id'].string_value == run_id and e.properties['component_id'].string_value == component_id],
            f"No {component_id} execution for context {context_id} and run {run_id} could be found"
        )

    def get_artifact_id(self, execution_id, step_name):
        request = metadata_store_service_pb2.GetEventsByExecutionIDsRequest()
        request.execution_ids.extend([execution_id])
        response = self.grpc_client.GetEventsByExecutionIDs(request)
    
        return first_or_raise(
            [e.artifact_id for e in response.events if any(step for step in e.path.steps if step.key == step_name)],
            f"No artifact for execution {execution_id} and step {step_name} could be found"
        )

    def get_artifact_property(self, artifact_id, property_name):
        artifact_request = metadata_store_service_pb2.GetArtifactsByIDRequest()
        artifact_request.artifact_ids.extend([artifact_id])
        artifact_response = self.grpc_client.GetArtifactsByID(artifact_request)
        return first_or_raise(
            [a.custom_properties[property_name].string_value for a in artifact_response.artifacts if property_name in a.custom_properties],
            f"No {property_name} for artifact {artifact_id} could be found"
        )

def first_or_raise(list, message):
    if list:
        return list[0]
    else:
        raise RuntimeError(message)
