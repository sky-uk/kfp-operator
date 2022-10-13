from typing import Optional

import kfp_server_api
from kfp.cli.cli import main as kfp_main
from kfp.cli.cli import cli as kfp_cli
from kfp._client import Client
import kfp_server_api
from .job import job

from typing import Optional

def delete_recurring_run(self, job_id):
    """Delete job.

    Args:
      job_id: id of the job.
    Returns:
      Object. If the method is called asynchronously, returns the request thread.
    Throws:
      Exception if experiment is not found.
    """
    return self._job_api.delete_job(id=job_id)

def list_pipeline_versions(
        self,
        resource_key_id: str,
        page_token: str = '',
        page_size: int = 10,
        sort_by: str = '',
        resource_key_type: Optional[str] = None,
        filter: Optional[str] = None
) -> kfp_server_api.ApiListPipelinesResponse:
    """Lists pipelines.
    Args:
        resource_key_type: The type of the resource that is referred to.
        resource_key_id: The ID of the resource that is referred to.
        page_token: Token for starting of the page.
        page_size: Size of the page.
        sort_by: one of 'field_name', 'field_name desc'. For example,
            'name desc'.
        filter: A url-encoded, JSON-serialized Filter protocol buffer
            (see [filter.proto](https://github.com/kubeflow/pipelines/blob/master/backend/api/filter.proto)).
            An example filter string would be:
                # For the list of filter operations please see:
                # https://github.com/kubeflow/pipelines/blob/master/sdk/python/kfp/_client.py#L40
                json.dumps({
                    "predicates": [{
                        "op": _FILTER_OPERATIONS["EQUALS"],
                        "key": "name",
                        "stringValue": "my-name",
                    }]
                })
    Returns:
        kfp_server_api.ApiListPipelineVersionsResponse: ApiListPipelineVersionsResponse object.
    """
    return self._pipelines_api.list_pipeline_versions(
        resource_key_type=resource_key_type,
        resource_key_id=resource_key_id,
        page_token=page_token,
        page_size=page_size,
        sort_by=sort_by,
        filter=filter)

def patch_client():
    setattr(Client, delete_recurring_run.__name__, delete_recurring_run)
    setattr(Client, list_pipeline_versions.__name__, list_pipeline_versions)

def main():
    patch_client()
    kfp_cli.add_command(job)
    kfp_main()
