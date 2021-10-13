import click

from kfp.cli.cli import main as kfp_main
from kfp.cli.cli import cli as kfp_cli
from .job import job

def main():
    kfp_cli.add_command(job)
    kfp_main()
