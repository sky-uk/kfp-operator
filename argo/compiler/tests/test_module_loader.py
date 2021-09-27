from compiler import compiler
import os
import sys
import pytest

def components_fn():
    return os.environ

def test_load_fn():
    env = {
        'a': 'aVal',
        'b': 'bVal'
    }
    
    sys.path.append(os.path.dirname(__file__))
    result = compiler.load_fn('test_module_loader.components_fn', env)()
    assert env.items() <= result.items()

def test_load_fn_invalid_fn():
    with pytest.raises(Exception) as e_info:
        compiler.load_fn('nonexistent.components_fn', {})()
