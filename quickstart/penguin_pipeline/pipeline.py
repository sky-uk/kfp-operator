### Environmental parameters can be left out when using the operator.
### Also, the return type is now a list of components instead of a pipeline.
#
#def create_pipeline(pipeline_name: str, pipeline_root: str, data_root: str,
#                    module_file: str, serving_model_dir: str,
#                    metadata_path: str) -> tfx.dsl.Pipeline:
def create_pipeline(data_root: str, module_file: str) -> list[tfx.dsl.components.base.base_node.BaseNode]:
    """Creates a penguin pipeline with TFX."""
    # Brings data into the pipeline.
    example_gen = tfx.components.CsvExampleGen(input_base=data_root)

    # Uses user-provided Python function that trains a model.
    trainer = tfx.components.Trainer(
        module_file=module_file,
        examples=example_gen.outputs['examples'],
        train_args=tfx.proto.TrainArgs(num_steps=100),
        eval_args=tfx.proto.EvalArgs(num_steps=5))

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
        example_gen,
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