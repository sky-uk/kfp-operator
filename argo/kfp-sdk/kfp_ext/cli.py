import click

from kfp.cli.cli import main as kfp_main
from kfp.cli.cli import cli as kfp_cli
from kfp._client import Client
from .job import job

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

def patch_client():
    setattr(Client, delete_recurring_run.__name__, delete_recurring_run)

def main():
    patch_client()
    kfp_cli.add_command(job)
    kfp_main()
