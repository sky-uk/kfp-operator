from grpc import insecure_channel
from ml_metadata.proto import metadata_store_pb2
from ml_metadata.proto import metadata_store_service_pb2
from ml_metadata.proto import metadata_store_service_pb2_grpc

channel = insecure_channel('localhost:8085')
stub = metadata_store_service_pb2_grpc.MetadataStoreServiceStub(channel)

context_request = metadata_store_service_pb2.GetContextByTypeAndNameRequest()
context_request.type_name = 'pipeline'
context_request.context_name = 'pipeline-sample'

context_response = stub.GetContextByTypeAndName(context_request)

context_id = context_response.context.id

request = metadata_store_service_pb2.GetExecutionsByContextRequest()
request.context_id = context_id
response = stub.GetExecutionsByContext(request)
for e in response.executions:
    if e.properties['run_id'].string_value == 'pipeline-sample-hjmpq':
        event_request = metadata_store_service_pb2.GetEventsByExecutionIDsRequest()
        event_request.execution_ids.extend([e.id])
        event_response = stub.GetEventsByExecutionIDs(event_request)
        for e in event_response.events:
            if e.path.steps[0].key == 'pushed_model':
                artifact_id = e.artifact_id
                artifact_request = metadata_store_service_pb2.GetArtifactsByIDRequest()
                artifact_request.artifact_ids.extend([artifact_id])
                artifact_response = stub.GetArtifactsByID(artifact_request)
                for a in artifact_response.artifacts:
                    print(a.custom_properties['pushed_destination'].string_value)
