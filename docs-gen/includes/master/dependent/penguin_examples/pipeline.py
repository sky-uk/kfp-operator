import os
from typing import List
from tfx.components import CsvExampleGen, Trainer
from tfx.proto import trainer_pb2
from tfx.dsl.components.base.base_node import BaseNode

def create_components() -> List[BaseNode]:
    """Creates a penguin pipeline with TFX."""
    # Brings data into the pipeline.
    example_gen = CsvExampleGen(input_base='/data')

    return [ example_gen ]
