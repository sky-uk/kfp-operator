from click.testing import CliRunner
from kfp.cli.output import OutputFormat
from kfp_ext.job import job, get
from unittest.mock import Mock
import unittest
import textwrap

class TestJob(unittest.TestCase):
    runner = CliRunner()
    experiment_name = 'an_experiment'
    experiment_id = '12345'
    job_name = 'a_job'
    job_id = '67890'
    pipeline_name = 'a_pipeline'
    pipeline_id = 'abcdef'
    cron_expression = '* * * * * *'

    @staticmethod
    def experiment_with_id(id):
        experiment = Mock()
        experiment.id = id
        return experiment

    @staticmethod
    def job(id, name):
        job = Mock()
        job.id = id
        job.name = name
        return job

    def test_get_table(self):
        client = Mock()
        client.get_recurring_run.return_value = self.job(self.job_id, self.job_name)

        result = self.runner.invoke(job, [
                'get',
                self.job_id
            ], obj={'client': client, 'output': OutputFormat.table.name})

        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.output, textwrap.dedent(
        f"""\
         Job Details
         -------------
         ID    {self.job_id}
         Name  {self.job_name}
         """))

    def test_get_json(self):
        client = Mock()
        client.get_recurring_run.return_value = self.job(self.job_id, self.job_name)

        result = self.runner.invoke(job, [
                'get',
                self.job_id
            ], obj={'client': client, 'output': OutputFormat.json.name})

        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.output, textwrap.dedent(
        f"""\
         {{
             "Job Details": {{
                 "ID": "{self.job_id}",
                 "Name": "{self.job_name}"
             }}
         }}
         """))

    def test_submit_job_without_cron(self):
        client = Mock()

        result = self.runner.invoke(job, [
                'submit',
                '--job-name', self.job_name,
                '--experiment-name', self.experiment_name,
                '--pipeline-name', self.pipeline_name
            ], obj={'client': client, 'output': None})

        self.assertEqual(result.exit_code, 2)
        self.assertIn("Error: Missing option '-c' / '--cron-expression'.", result.output)

    def test_submit_job_without_job(self):
        client = Mock()

        result = self.runner.invoke(job, ['submit'], obj={'client': client, 'output': None})

        self.assertEqual(result.exit_code, 2)
        self.assertIn("Error: Missing option '-j' / '--job-name'.", result.output)

    def test_submit_job_without_experiment(self):
        client = Mock()

        result = self.runner.invoke(job, [
                'submit',
                '--job-name', self.job_name
            ], obj={'client': client, 'output': None})

        self.assertEqual(result.exit_code, 2)
        self.assertIn("Error: Missing option '-e' / '--experiment-name'.", result.output)

    def test_submit_job_without_pipeline(self):
        client = Mock()

        result = self.runner.invoke(job, [
                'submit',
                '--job-name', self.job_name,
                '--experiment-name', self.experiment_name,
                '--cron-expression', self.cron_expression
            ], obj={'client': client, 'output': None})

        self.assertEqual(result.exit_code, 1)
        self.assertEqual(result.output, 'You must provide one of [pipeline_name, pipeline_id].\n')

    def test_submit_job_with_pipeline_name(self):
        client = Mock()
        client.create_experiment.return_value = self.experiment_with_id(self.experiment_id)
        client.get_pipeline_id.return_value = self.pipeline_id
        client.create_recurring_run.return_value = self.job(self.job_id, self.job_name)

        result = self.runner.invoke(job, [
                'submit',
                '--job-name', self.job_name,
                '--pipeline-name', self.pipeline_name,
                '--experiment-name', self.experiment_name,
                '--cron-expression', self.cron_expression
            ], obj={'client': client, 'output': OutputFormat.table.name})


        client.create_recurring_run.assert_called_once_with(experiment_id=self.experiment_id, job_name=self.job_name, cron_expression=self.cron_expression, pipeline_id=self.pipeline_id)
        self.assertEqual(result.exit_code, 0)
        # Output format is tested in test_get_*
        self.assertIn(self.job_id, result.output)
        self.assertIn(self.job_name, result.output)

    def test_submit_job_with_pipeline_id(self):
        client = Mock()
        client.create_experiment.return_value = self.experiment_with_id(self.experiment_id)
        client.create_recurring_run.return_value = self.job(self.job_id, self.job_name)

        result = self.runner.invoke(job, [
                'submit',
                '--job-name', self.job_name,
                '--pipeline-id', self.pipeline_id,
                '--experiment-name', self.experiment_name,
                '--cron-expression', self.cron_expression
            ], obj={'client': client, 'output': OutputFormat.table.name})

        client.create_recurring_run.assert_called_once_with(experiment_id=self.experiment_id, job_name=self.job_name, cron_expression=self.cron_expression, pipeline_id=self.pipeline_id)
        self.assertEqual(result.exit_code, 0)
        # Output format is tested in test_get_*
        self.assertIn(self.job_id, result.output)
        self.assertIn(self.job_name, result.output)

    def test_delete_job_without_id(self):
        client = Mock()

        result = self.runner.invoke(job, [
                'delete',
            ], obj={'client': client, 'output': None})

        self.assertEqual(result.exit_code, 2)
        self.assertIn("Missing argument 'JOB_ID'", result.output)

    def test_delete_job_with_id(self):
        client = Mock()

        result = self.runner.invoke(job, [
                'delete',
                self.job_id,
            ], obj={'client': client, 'output': OutputFormat.table.name})

        client.delete_recurring_run.assert_called_once_with(job_id=self.job_id)
        self.assertEqual(result.exit_code, 0)
        self.assertEqual(result.output, f"{self.job_id} is deleted.\n")