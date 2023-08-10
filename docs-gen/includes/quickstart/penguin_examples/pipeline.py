import os
from typing import List
from tfx.components import CsvExampleGen, Trainer
from tfx.proto import trainer_pb2
from tfx.dsl.components.base.base_node import BaseNode

### Environmental parameters can be left out when using the operator.
### Also, the return type is now a list of components instead of a pipeline.
#
#def create_pipeline(pipeline_name: str, pipeline_root: str, data_root: str,
#                    module_file: str, serving_model_dir: str,
#                    metadata_path: str) -> tfx.dsl.Pipeline:
def create_components() -> List[BaseNode]:
    """Creates a penguin pipeline with TFX."""
    # Brings data into the pipeline.
    example_gen = CsvExampleGen(input_base='/data')

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
        example_gen
    ]

    return components
