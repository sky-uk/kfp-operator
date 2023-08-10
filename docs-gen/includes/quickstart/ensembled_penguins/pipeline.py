import os
from typing import List
from tfx.components import CsvExampleGen, Trainer, ImportExampleGen
from tfx.proto import trainer_pb2, example_gen_pb2
from tfx.dsl.components.base.base_node import BaseNode
from tfx.v1.dsl.experimental import RuntimeParameter
from typing import Text

### Environmental parameters can be left out when using the operator.
### Also, the return type is now a list of components instead of a pipeline.
#
#def create_pipeline(pipeline_name: str, pipeline_root: str, data_root: str,
#                    module_file: str, serving_model_dir: str,
#                    metadata_path: str) -> tfx.dsl.Pipeline:
def create_components() -> List[BaseNode]:
    """Creates a penguin pipeline with TFX."""

    # Defines a pipeline runtime parameter
    examples_location_param = RuntimeParameter(
        name = "examples_location",
        ptype = Text
    )

    # Brings data into the pipeline.
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

    # Uses user-provided Python function that trains a model.
    trainer = Trainer(
        run_fn='trainer.run_fn',
        examples=examples.outputs['examples'],
        train_args=trainer_pb2.TrainArgs(num_steps=100),
        eval_args=trainer_pb2.EvalArgs(num_steps=5))

    ### This needs to be omitted when using the operator.
    #
    ## Pushes the model to a filesystem destination.
    #pusher = tfx.components.Pusher(
    #  model=trainer.outputs['model'],
    #  push_destination=tfx.proto.PushDestination(
    #      filesystem=tfx.proto.PushDestination.Filesystem(
    #          base_directory=serving_model_dir)))

    # Following three components will be included in the pipeline.
    components = [
        examples,
        trainer,
        #pusher
    ]

    ### When using the operator, it creates the pipeline for us,
    ### so we return the components directly instead.
    #
    #return tfx.dsl.Pipeline(
    #  pipeline_name=pipeline_name,
    #  pipeline_root=pipeline_root,
    #  metadata_connection_config=tfx.orchestration.metadata
    #  .sqlite_metadata_connection_config(metadata_path),
    #  components=components)

    return components
