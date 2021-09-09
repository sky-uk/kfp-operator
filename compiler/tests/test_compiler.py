from compiler import compiler

def test_dict_to_cli_args():
    args = {
        'a': 'aVal',
        'b': 'bVal'
    }

    print(compiler.dict_to_cli_args(args))
    assert compiler.dict_to_cli_args(args) == [
        '--a=aVal',
        '--b=bVal'
    ]