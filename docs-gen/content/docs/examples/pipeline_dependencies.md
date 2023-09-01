---
title: "Pipeline Dependencies"
weight: 2
---

Pipeline dependencies allow splitting up larger machine learning pipelines into sub-pipelines. This is particularly useful when:
- The data of an earlier step changes at a lower frequency than the data for subsequent steps
- Outputs of an earlier step could be shared between pipelines to avoid re-processing the same data

In this example, we break up the penguin example pipeline into two pipelines:
- The *Penguin Examples Pipeline* has a single pipeline step that imports the CSV example and outputs it as an artifact
- The *Penguin Training Pipeline* references the previously produces example and trains the model 

![ensembling.svg](/images/ensembling.svg)

The examples for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/dependent).

## 1. Build and deploy the Penguin Examples Pipeline

The penguin examples pipeline loads the CSV data and makes it available as an output artifact.

First, create the pipeline code in `penguin_examples/pipeline.py`:

```python
{{% readfile file="includes/dependent/penguin_examples/pipeline.py" %}}
```

Next, create `penguin_examples/Dockerfile`, then build and push the pipeline image:

```dockerfile
{{% readfile file="includes/dependent/penguin_examples/Dockerfile" %}}
```

```bash
docker build penguin_examples -t kfp-dependent-examples:v1
docker push kfp-dependent-examples:v1
```

Now create and apply the resources needed to compile and train the penguin examples pipeline:

Create `pipeline_examples.yaml`. The specification has not changed from the original example.

```yaml
{{% readfile file="includes/dependent/resources/pipeline_examples.yaml" %}}
```

Create `runconfiguration_examples.yaml` which schedules the pipeline to be trained at regular intervals and specifies the output artifacts that it exposes.
```yaml
{{% readfile file="includes/dependent/resources/runconfiguration_examples.yaml" %}}
```

Finally, apply the resources:

```bash
kubectl apply -f resources/pipeline_examples.yaml
kubectl apply -f resources/runconfiguration_examples.yaml
```

## 2. Build and deploy the Penguin Training Pipeline

Create `penguin_training/pipeline.py` which imports the artifact as referenced by its runtime parameters. Note that the `CsvExampleGen` has been replaced by a `ImportExampleGen`:

```python
{{% readfile file="includes/dependent/penguin_training/pipeline.py" %}}
```

Next create `penguin_training/trainer.py` which has not changed from the original pipeline:

```python
{{% readfile file="includes/dependent/penguin_training/trainer.py" %}}
```

Create `penguin_training/Dockerfile`. In contrast to the above, this Dockerfile does not include the examples CSV.

```dockerfile
{{% readfile file="includes/dependent/penguin_training/Dockerfile" %}}
```

Next, build the pipeline as a Docker container and push it:

```bash
docker build penguin_training -t kfp-dependent-training:v1
docker push kfp-dependent-training:v1
```

Now create and apply the resources needed to compile and train the penguin training pipeline:

Create `pipeline_training.yaml`. The specification has not changed from the original example.

```yaml
{{% readfile file="includes/dependent/resources/pipeline_training.yaml" %}}
```

Create `runconfiguration_training.yaml`. A run will be triggered as soon as the dependency has finished training and produced the referenced artifact. This artifact will then be provided to the pipeline as a runtime parameter.

```yaml
{{% readfile file="includes/dependent/resources/runconfiguration_training.yaml" %}}
```

Apply the resources as follows:

```bash
kubectl apply -f resources/pipeline_training.yaml
kubectl apply -f resources/runconfiguration_training.yaml
```
