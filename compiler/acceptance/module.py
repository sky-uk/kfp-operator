from typing import List
import tensorflow as tf

from tfx import v1 as tfx


# TFX Trainer will call this function.
def run_fn(fn_args: tfx.components.FnArgs):
    print('doing nnothing')