import importlib
import os

def load_fn(tfx_components, env):
    for key, value in env.items():
        os.environ[key] = value

    (module_name, fn_name) = tfx_components.rsplit('.', 1)
    module = importlib.import_module(module_name)
    fn = getattr(module, fn_name)

    return fn