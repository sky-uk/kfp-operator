---
title: "Pipeline Training"
weight: 1
---

This tutorial walks you through the creation of a simple TFX pipeline on Kubeflow Pipelines and shows you how to manage pipelines via Kubernetes Custom Resources.

The examples for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.6.0/quickstart).

## 1. Build the Pipeline

We use the same pipeline as the [TFX example](https://www.tensorflow.org/tfx/tutorials/tfx/penguin_simple) with a few modifications.

Create `pipeline.py`.
Note that the pipeline definition itself is simpler because all infrastructure references, like pusher and pipeline root, will be injected by the operator before the pipeline is uploaded to Kubeflow:

```python
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

    # Uses user-provided Python function that trains a model.
    trainer = Trainer(
        run_fn='trainer.run_fn',
        examples=example_gen.outputs['examples'],
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
```

Create `trainer.py`.
The training code remains unchanged:

```python
from typing import List
from absl import logging
import tensorflow as tf
from tensorflow import keras
from tensorflow_transform.tf_metadata import schema_utils

from tfx import v1 as tfx
from tfx_bsl.public import tfxio
from tensorflow_metadata.proto.v0 import schema_pb2

_FEATURE_KEYS = [
    'culmen_length_mm', 'culmen_depth_mm', 'flipper_length_mm', 'body_mass_g'
]
_LABEL_KEY = 'species'

_TRAIN_BATCH_SIZE = 20
_EVAL_BATCH_SIZE = 10

# Since we're not generating or creating a schema, we will instead create
# a feature spec.  Since there are a fairly small number of features this is
# manageable for this dataset.
_FEATURE_SPEC = {
    **{
        feature: tf.io.FixedLenFeature(shape=[1], dtype=tf.float32)
           for feature in _FEATURE_KEYS
       },
    _LABEL_KEY: tf.io.FixedLenFeature(shape=[1], dtype=tf.int64)
}


def _input_fn(file_pattern: List[str],
              data_accessor: tfx.components.DataAccessor,
              schema: schema_pb2.Schema,
              batch_size: int = 200) -> tf.data.Dataset:
  """Generates features and label for training.

  Args:
    file_pattern: List of paths or patterns of input tfrecord files.
    data_accessor: DataAccessor for converting input to RecordBatch.
    schema: schema of the input data.
    batch_size: representing the number of consecutive elements of returned
      dataset to combine in a single batch

  Returns:
    A dataset that contains (features, indices) tuple where features is a
      dictionary of Tensors, and indices is a single Tensor of label indices.
  """
  return data_accessor.tf_dataset_factory(
      file_pattern,
      tfxio.TensorFlowDatasetOptions(
          batch_size=batch_size, label_key=_LABEL_KEY),
      schema=schema).repeat()


def _build_keras_model() -> tf.keras.Model:
  """Creates a DNN Keras model for classifying penguin data.

  Returns:
    A Keras Model.
  """
  # The model below is built with Functional API, please refer to
  # https://www.tensorflow.org/guide/keras/overview for all API options.
  inputs = [keras.layers.Input(shape=(1,), name=f) for f in _FEATURE_KEYS]
  d = keras.layers.concatenate(inputs)
  for _ in range(2):
    d = keras.layers.Dense(8, activation='relu')(d)
  outputs = keras.layers.Dense(3)(d)

  model = keras.Model(inputs=inputs, outputs=outputs)
  model.compile(
      optimizer=keras.optimizers.Adam(1e-2),
      loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
      metrics=[keras.metrics.SparseCategoricalAccuracy()])

  model.summary(print_fn=logging.info)
  return model


# TFX Trainer will call this function.
def run_fn(fn_args: tfx.components.FnArgs):
  """Train the model based on given args.

  Args:
    fn_args: Holds args used to train the model as name/value pairs.
  """

  # This schema is usually either an output of SchemaGen or a manually-curated
  # version provided by pipeline author. A schema can also derived from TFT
  # graph if a Transform component is used. In the case when either is missing,
  # `schema_from_feature_spec` could be used to generate schema from very simple
  # feature_spec, but the schema returned would be very primitive.
  schema = schema_utils.schema_from_feature_spec(_FEATURE_SPEC)

  train_dataset = _input_fn(
      fn_args.train_files,
      fn_args.data_accessor,
      schema,
      batch_size=_TRAIN_BATCH_SIZE)
  eval_dataset = _input_fn(
      fn_args.eval_files,
      fn_args.data_accessor,
      schema,
      batch_size=_EVAL_BATCH_SIZE)

  model = _build_keras_model()
  model.fit(
      train_dataset,
      steps_per_epoch=fn_args.train_steps,
      validation_data=eval_dataset,
      validation_steps=fn_args.eval_steps)

  # The result of the training should be saved in `fn_args.serving_model_dir`
  # directory.
  model.save(fn_args.serving_model_dir, save_format='tf')
```

Create `Dockerfile`.

```dockerfile
FROM python:3.10.12 as builder

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

RUN pip install poetry==1.7.1

COPY pyproject.toml .
COPY poetry.lock .

RUN poetry config virtualenvs.create true
RUN poetry config virtualenvs.in-project true
RUN poetry install --no-root

FROM python:3.10.12-slim as runtime

ENV VIRTUAL_ENV=/.venv \
    PATH="/.venv/bin:$PATH"

WORKDIR /pipeline

COPY --from=builder ${VIRTUAL_ENV} ${VIRTUAL_ENV}
COPY --from=builder /data /data

COPY penguin_pipeline/*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
```

Next, build the pipeline as a Docker container and push it:

```bash
docker build . -t kfp-quickstart:v1
...
docker push kfp-quickstart:v1
```

## 2. Create a Pipeline Resource

Now that we have a pipeline image, we can create a `pipeline.yaml` resource to manage the lifecycle of this pipeline on Kubeflow:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Pipeline
metadata:
  name: penguin-pipeline
spec:
  image: kfp-quickstart:v1
  tfxComponents: pipeline.create_components
```

```bash
kubectl apply -f resources/pipeline.yaml
```

The pipeline now gets uploaded to Kubeflow in several steps. After a few seconds to minutes, the following command should result in a success:

```bash
kubectl get pipeline

NAME               SYNCHRONIZATIONSTATE   PROVIDERID
penguin-pipeline   Succeeded              53905abe-0337-48de-875d-67b9285f3cf7
```

Now visit your Kubeflow Pipelines UI. You should be able to see the newly created pipeline named `penguin-pipeline`. Note that you will see two versions: 'penguin-pipeline' and 'v1'. This is due to an [open issue on Kubeflow](https://github.com/kubeflow/pipelines/issues/5881) where you can't specify a version when creating a pipeline.

## 3. Create an Experiment resource

Note: this step is optional. You can continue with the next step if you want to use the `Default` experiment instead.

Create `experiment.yaml`:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Experiment
metadata:
  name: penguin-experiment
spec:
  description: An experiment for the penguin pipeline
```

```bash
kubectl apply -f resources/experiment.yaml
```

## 4. Create a pipeline RunConfiguration resource

We can now define a recurring run declaratively using the `RunConfiguration` resource.

Note: remove `experimentName` if you want to use the `Default` experiment instead of `penguin-experiment`

Create `runconfiguration.yaml`:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: RunConfiguration
metadata:
    name: penguin-pipeline-recurring-run
spec:
    run:
      pipeline: penguin-pipeline
      experimentName: penguin-experiment
    triggers:
      schedules:
      - cronExpression: 0 * * * *
        startTime: 2024-01-01T00:00:00Z
        endTime: 2024-12-31T23:59:59Z
```

```bash
kubectl apply -f resources/runconfiguration.yaml
```

This will trigger run of `penguin-pipeline` once every hour. Note that the cron schedule uses a 6-place space separated syntax as defined [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).

## 5. (Optional) Deploy newly trained models

If the operator has been installed with [Argo-Events](https://argoproj.github.io/argo-events/) support, we can now specify eventsources and sensors to update arbitrary Kubernetes config when a pipeline has been trained successfully.
In this example we are updating a serving component with the location of the newly trained model. 

Create `apply-model-location.yaml`. This creates an `EventSource` and a `Sensor` as well as an `EventBus`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: EventBus
metadata:
  name: default
spec:
  nats:
    native: {}
---
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: run-completion
spec:
  nats:
    run-completion:
      jsonBody: true
      subject: events
      url: nats://eventbus-kfp-operator-events-stan-svc.kfp-operator.svc:4222
---
apiVersion: argoproj.io/v1alpha1
kind: Sensor
metadata:
  name: penguin-pipeline-model-update
spec:
  template:
    serviceAccountName: events-sa
  dependencies:
    - name: run-completion
      eventSourceName: run-completion
      eventName: run-completion
      filters:
        data:
          - path: body.data.status
            type: string
            comparator: "="
            value:
              - "succeeded"
          - path: body.data.pipelineName
            type: string
            comparator: "="
            value:
              - "penguin-pipeline"
  triggers:
    - template:
        name: update
        k8s:
          operation: update
          source:
            resource:
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: serving-config
              data:
                servingModel: ""
          parameters:
            - src:
                dependencyName: run-completion
                dataKey: body.data.artifacts.0.location
              dest: data.servingModel
```

```bash
kubectl apply -f resources/apply-model-location.yaml
```
