from compiler import compiler
import os
import sys
import pytest


def components_fn():
    return os.environ


def test_load_fn():
    env = [
        {'name': 'a', 'value': 'aVal'},
        {'name': 'a', 'value': 'overriddenVal'},
        {'name': 'b', 'value': 'bVal'},
    ]
    
    sys.path.append(os.path.dirname(__file__))
    result = compiler.load_fn('test_module_loader.components_fn', env)()

    assert result['a'] == 'overriddenVal'
    assert result['b'] == 'bVal'


def test_load_fn_invalid_fn():
    with pytest.raises(Exception) as e_info:
        compiler.load_fn('nonexistent.components_fn', {})()


def test_load_fn_invalid_module_path():
    with pytest.raises(Exception) as e_info:
        compiler.load_fn('components_fn', {})()

   assert str(e_info.value) == "Invalid pipeline format: [components_fn]. Expected format: 'module_path.function_name'."
