from compiler import compiler


def test_name_values_to_cli_args():
    name_values = [
        {"name": "aName", "value": "aValue"},
        {"name": "aName", "value": "aValue2"},
        {"name": "anotherName", "value": "anotherValue"}
    ]

    assert compiler.name_values_to_cli_args(name_values) == [
        '--aName=aValue',
        '--aName=aValue2',
        '--anotherName=anotherValue',
    ]


def test_pipeline_paths_for_config():
    pipeline_config = {'name': 'pipeline'}
    provider_config = {'pipelineRootStorage': "pipeline_root"}

    pipeline_root, temp_directory = compiler.pipeline_paths_for_config(pipeline_config, provider_config)

    assert pipeline_root == "pipeline_root/pipeline"
    assert temp_directory == "pipeline_root/pipeline/tmp"


def test_sanitise_namespaced_pipeline_name():
    assert compiler.sanitise_namespaced_pipeline_name("pipeline-name") == "pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("/pipeline-name") == "-pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("mlops/pipeline-name") == "mlops-pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("") == ""
