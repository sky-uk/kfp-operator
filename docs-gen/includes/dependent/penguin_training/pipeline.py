import os
from typing import List
from tfx.components import CsvExampleGen, Trainer, ImportExampleGen
from tfx.proto import trainer_pb2, example_gen_pb2
from tfx.dsl.components.base.base_node import BaseNode
from tfx.v1.dsl.experimental import RuntimeParameter
from typing import Text

def create_components() -> List[BaseNode]:
    """Creates a penguin pipeline with TFX."""

    # Defines a pipeline runtime parameter
    examples_location_param = RuntimeParameter(
        name = "examples_location",
        ptype = Text
    )

    # Imports the artifact referenced by the runtime parameter
    examples = ImportExampleGen(
        input_base = examples_location_param,
        input_config=example_gen_pb2.Input(
              splits=[
                  example_gen_pb2.Input.Split(
                      name="eval",
                      pattern='Split-eval/*'
                      ),
                  example_gen_pb2.Input.Split(
                      name="train",
                      pattern='Split-train/*'
                      ),
              ]
            ),
    )

    trainer = Trainer(
        run_fn='trainer.run_fn',
        examples=examples_location_param,
        train_args=trainer_pb2.TrainArgs(num_steps=100),
        eval_args=trainer_pb2.EvalArgs(num_steps=5))

    components = [
        examples,
        trainer,
    ]

    return components
