import click
from kfp.cli.output import print_output, OutputFormat

@click.group()
def job():
    """manage job resources."""
    pass

@job.command()
@click.option(
    '-j',
    '--job-name',
    required=True,
    help='Name of the Job.')
@click.option('-p', '--pipeline-id', help='ID of the pipeline template.')
@click.option('-n', '--pipeline-name', help='Name of the pipeline template.')
@click.option(
    '-e',
    '--experiment-name',
    required=True,
    help='Experiment name of the run.')
@click.option(
    '-e',
    '--experiment-name',
    required=True,
    help='Experiment name of the run.')
@click.option(
    '-c',
    '--cron-expression',
    required=True,
    help='Cron Expression of the run.')
@click.pass_context
def submit(ctx, job_name, pipeline_id, pipeline_name, experiment_name, cron_expression):
    client = ctx.obj['client']
    output_format = ctx.obj['output']

    experiment = client.create_experiment(experiment_name)

    if not pipeline_id:
        if pipeline_name:
            pipeline_id = client.get_pipeline_id(name=pipeline_name)
        else:
            click.echo(
                'You must provide one of [pipeline_name, pipeline_id].',
                err=True)
            sys.exit(1)

    job = client.create_recurring_run(
        experiment_id=experiment.id,
        job_name=job_name,
        cron_expression=cron_expression,
        pipeline_id=pipeline_id)

    _display_job(job, output_format)

@job.command()
@click.argument('job-id')
@click.pass_context
def get(ctx, job_id):
    client = ctx.obj['client']
    output_format = ctx.obj['output']

    job = client.get_recurring_run(job_id)

    _display_job(job, output_format)

@job.command()
@click.argument('job-id')
@click.pass_context
def delete(ctx, job_id):
    client = ctx.obj['client']

    client.delete_recurring_run(job_id=job_id)
    click.echo("{} is deleted.".format(job_id))

def _display_job(job, output_format):
    table = [
        ["ID", job.id],
        ["Name", job.name]
    ]

    if output_format == OutputFormat.table.name:
        print_output([], ["Job Details"], output_format)
        print_output(table, [], output_format, table_format="plain")
    elif output_format == OutputFormat.json.name:
        output = dict()
        output["Job Details"] = dict(table)
        print_output(output, [], output_format)
