from kfp_compiler import compiler


def test_dict_to_cli_args():
    args = {
        'a': ['aVal'],
        'b': ['bVal'],
        'c': ['cVal1', 'cVal2'],
    }

    assert compiler.dict_to_cli_args(args) == [
        '--a=aVal',
        '--b=bVal',
        '--c=cVal1',
        '--c=cVal2',
    ]


def test_pipeline_paths_for_config():
    pipeline_config = {'name': 'pipeline'}
    provider_config = {'pipelineRootStorage': "pipeline_root"}

    pipeline_root, serving_model_directory, temp_directory = compiler.pipeline_paths_for_config(pipeline_config, provider_config)

    assert pipeline_root == "pipeline_root/pipeline"
    assert serving_model_directory == "pipeline_root/pipeline/serving"
    assert temp_directory == "pipeline_root/pipeline/tmp"


def test_merge_multimap():
    multimap1 = {'someArgName1': ['someArgValue1']}
    multimap2 = {'defaultArgName1': ['defaultArgValue1'], 'someArgName1': ['someArgValue2']}

    merged = compiler.merge_multimap(multimap1, multimap2)

    assert merged == {'someArgName1': ['someArgValue1', 'someArgValue2'], 'defaultArgName1': ['defaultArgValue1']}
