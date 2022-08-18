from kfp_compiler import compiler

def test_dict_to_cli_args():
    args = {
        'a': 'aVal',
        'b': 'bVal',
        'c': ['cVal1', 'cVal2'],
    }

    assert compiler.dict_to_cli_args(args) == [
        '--a=aVal',
        '--b=bVal',
        '--c=cVal1',
        '--c=cVal2',
    ]
