---
title: "Pipeline Dependencies"
weight: 2
---

Pipeline dependencies allow splitting up larger machine learning pipelines into sub-pipelines. This is particularly useful when:
- The data of an earlier step changes at a lower frequency than the data for subsequent steps
- Outputs of an earlier step could be shared between pipelines to avoid re-processing the same data
- Serving a combined model of two or more pipelines

In this example, we break up the penguin example pipeline into two pipelines:
- The *Penguin Examples Pipeline* has a single pipeline step that imports the CSV example and outputs it as an artifact
- The *Penguin Training Pipeline* references the previously produced example and trains the model

![ensembling.svg]({{< param "subpath" >}}/versions/v0.6.0/images/ensembling.svg)

This example might not be of practical use but it demonstrates the approach to managing pipeline dependencies.

The code for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.6.0/dependent).

## 1. Build and deploy the Penguin Examples Pipeline

The penguin examples pipeline loads the CSV data and makes it available as an output artifact.

First, create the pipeline code in `penguin_examples/pipeline.py`:

```python
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
```

Next, create `penguin_examples/Dockerfile`, then build and push the pipeline image:

```Dockerfile
# TFX build
FROM tensorflow/tfx:1.13.0

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

WORKDIR /pipeline
COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
```

```bash
docker build penguin_examples -t kfp-dependent-examples:v1
docker push kfp-dependent-examples:v1
```

Now create and apply the resources needed to compile and train the penguin examples pipeline:

Create `pipeline_examples.yaml`. The specification has not changed from the original example.

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Pipeline
metadata:
  name: penguin-pipeline-examples
spec:
  image: kfp-dependent-examples:v1
  tfxComponents: pipeline.create_components
```

Create `runconfiguration_examples.yaml` which schedules the pipeline to be trained at regular intervals and specifies the output artifacts that it exposes.

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: RunConfiguration
metadata:
  name: penguin-pipeline-examples-rc
spec:
  run:
    pipeline: penguin-pipeline-examples
    artifacts:
    - name: examples
      path: CsvExampleGen:examples
  triggers:
    schedules:
    - cronExpression: '0 * * * *'
      startTime: "2024-01-01T00:00:00Z"
      endTime: "2024-12-31T23:59:59Z"
```

Finally, apply the resources:

```bash
kubectl apply -f resources/pipeline_examples.yaml
kubectl apply -f resources/runconfiguration_examples.yaml
```

## 2. Build and deploy the Penguin Training Pipeline

Create `penguin_training/pipeline.py` which imports the artifact as referenced by its runtime parameters. Note that the `CsvExampleGen` has been replaced by a `ImportExampleGen`:

```python
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
```

Next create `penguin_training/trainer.py`, which has not changed from [the original pipeline](../pipeline_training/).

Create `penguin_training/Dockerfile`. In contrast to the above, this Dockerfile does not include the examples CSV.


```Dockerfile
# TFX build
FROM tensorflow/tfx:1.13.0

WORKDIR /pipeline
COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
```


Next, build the pipeline as a Docker container and push it:

```bash
docker build penguin_training -t kfp-dependent-training:v1
docker push kfp-dependent-training:v1
```

Now create and apply the resources needed to compile and train the penguin training pipeline.
`pipeline_training.yaml` has not changed from the original example:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Pipeline
metadata:
  name: penguin-pipeline-training
spec:
  image: kfp-dependent-training:v1
  tfxComponents: pipeline.create_components
```


`runconfiguration_training.yaml` has been updated so that a run is triggered as soon as the dependency has finished training and produced the referenced artifact. This artifact will then be provided to the pipeline as a runtime parameter:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: RunConfiguration
metadata:
  name: penguin-pipeline-training-rc
spec:
  run:
    pipeline: penguin-pipeline-training
    runtimeParameters:
    - name: examples_location
      valueFrom:
        runConfigurationRef:
          name: penguin-pipeline-examples-rc
          outputArtifact: examples
  triggers:
    runConfigurations:
      - penguin-pipeline-examples-rc
```

Apply the resources as follows:

```bash
kubectl apply -f resources/pipeline_training.yaml
kubectl apply -f resources/runconfiguration_training.yaml
```
