from compiler import compiler

def test_dict_to_cli_args():
    args = {
        'a': 'aVal',
        'b': 'bVal'
    }

    assert compiler.dict_to_cli_args(args) == [
        '--a=aVal',
        '--b=bVal'
    ]
