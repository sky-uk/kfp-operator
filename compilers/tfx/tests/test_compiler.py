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


def test_sanitise_namespaced_pipeline_name():
    assert compiler.sanitise_namespaced_pipeline_name("pipeline-name") == "pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("/pipeline-name") == "-pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("mlops/pipeline-name") == "mlops-pipeline-name"
    assert compiler.sanitise_namespaced_pipeline_name("") == ""
