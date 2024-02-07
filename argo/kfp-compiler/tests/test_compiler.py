from kfp_compiler import compiler


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

    pipeline_root, serving_model_directory, temp_directory = compiler.pipeline_paths_for_config(pipeline_config, provider_config)

    assert pipeline_root == "pipeline_root/pipeline"
    assert serving_model_directory == "pipeline_root/pipeline/serving"
    assert temp_directory == "pipeline_root/pipeline/tmp"
