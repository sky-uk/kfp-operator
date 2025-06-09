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
- The *Penguin Training Pipeline* references the previously produces example and trains the model

![ensembling.svg]({{< param "subpath" >}}/versions/v0.7.0/images/ensembling.svg)

This example might not be of practical use but it demonstrates the approach to managing pipeline dependencies.

The code for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.7.0/quickstart-with-dependencies).

## 1. Build and deploy the Penguin Examples Pipeline

The `penguin_examples` pipeline loads the CSV data and makes it available as an output artifact. The code for this pipeline can be found under [quickstart-base]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-base).

First, create the pipeline code in `penguin_examples/pipeline.py`:

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-base/penguin_examples/pipeline.py" code="true" lang="python" %}}

Next, create the following `Dockerfile`, then build and push the pipeline image:

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-base/Dockerfile" code="true" lang="Dockerfile" %}}

```bash
docker build quickstart-base -t kfp-quickstart-base:v1
docker push kfp-quickstart-base:v1
```

Now create and apply the resources needed to compile and train the penguin examples pipeline:

Create `quickstart-base/resources/pipeline.yaml`. The specification has not changed from the original example.

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-base/resources/pipeline.yaml" code="true" lang="yaml"%}}

Create `quickstart-base/resources/runconfiguration.yaml` which schedules the pipeline to be trained at regular intervals and specifies the output artifacts that it exposes.

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-base/resources/runconfiguration.yaml" code="true" lang="yaml"%}}

Finally, apply the resources:

```bash
kubectl apply -f quickstart-base/resources/pipeline.yaml
kubectl apply -f quickstart-base/resources/runconfiguration.yaml
```

## 2. Build and deploy the Penguin Training Pipeline

The `penguin_training` pipeline imports the artifact exposed by the `penguin_examples` pipeline and trains a model in the same way as [the original pipeline](../pipeline_training/).
The code for this pipeline can be found under [quickstart-dependant]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-dependant).

Create `penguin_training/pipeline.py` which imports the artifact as referenced by its runtime parameters. Note that the `CsvExampleGen` has been replaced by a `ImportExampleGen`:

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-dependant/penguin_training/pipeline.py" code="true" lang="python"%}}

Next create `penguin_training/trainer.py`, which has not changed from [the original pipeline](../pipeline_training/).

Create the following `Dockerfile`. In contrast to the above, this Dockerfile does not include the examples CSV.


{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-dependant/Dockerfile" code="true" lang="Dockerfile"%}}


Next, build the pipeline as a Docker container and push it:

```bash
docker build quickstart-dependant -t kfp-quickstart-dependant:v1
docker push kfp-quickstart-dependant:v1
```

Now create and apply the resources needed to compile and train the penguin training pipeline.
`quickstart-dependant/resources/pipeline.yaml` has not changed from the original example:

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-dependant/resources/pipeline.yaml" code="true" lang="yaml"%}}


`quickstart-dependant/resources/runconfiguration.yaml` has been updated so that a run is triggered as soon as the dependency has finished training and produced the referenced artifact. This artifact will then be provided to the pipeline as a runtime parameter:

{{% readfile file="/includes/versions/v0.7.0/quickstart-with-dependencies/quickstart-dependant/resources/runconfiguration.yaml" code="true" lang="yaml"%}}

Apply the resources as follows:

```bash
kubectl apply -f quickstart-dependant/resources/pipeline.yaml
kubectl apply -f quickstart-dependant/resources/runconfiguration.yaml
```
