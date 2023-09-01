---
title: "Pipeline Dependencies"
weight: 2
---

Pipeline dependencies allow splitting up larger machine learning pipelines into sub-pipelines. This is particularly useful when:
- The data of an earlier step changes at a lower frequency than the data for subsequent steps
- Outputs of an earlier step could be shared between pipelines to avoid re-processing the same data

In this example, we break up the dependent example into two pipelines:
- The *Examples Pipeline* has a single pipeline step that imports the CSV example and outputs it as an artifact
- The *Training Pipeline* references the previously produces example and trains the model 

![ensembling.svg](/images/ensembling.svg)

The examples for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/dependent).

## 1. Build the Pipelines

Create `penguin_examples/pipeline.py`.

```python
{{% readfile file="includes/dependent/penguin_examples/pipeline.py" %}}
```

Create `penguin_training/pipeline.py`.

```python
{{% readfile file="includes/dependent/penguin_training/pipeline.py" %}}
```

Create `penguin_examples/Dockerfile`.

```dockerfile
{{% readfile file="includes/dependent/penguin_examples/Dockerfile" %}}
```

Create `penguin_training/Dockerfile`.

```dockerfile
{{% readfile file="includes/dependent/penguin_training/Dockerfile" %}}
```

Next, build the pipeline as a Docker container and push it:

```bash
docker build penguin_examples -t kfp-dependent-examples:v1
docker build penguin_training -t kfp-dependent-training:v1
...
docker push kfp-dependent-examples:v1
docker push kfp-dependent-training:v1
```

## 2. Create Pipeline Resources

Create `pipeline_examples.yaml`:

```yaml
{{% readfile file="includes/dependent/resources/pipeline_examples.yaml" %}}
```

```bash
kubectl apply -f resources/pipeline_examples.yaml
```

Create `pipeline_training.yaml`:

```yaml
{{% readfile file="includes/dependent/resources/pipeline_training.yaml" %}}
```

```bash
kubectl apply -f resources/pipeline_training.yaml
```

## 4. Create dependent RunConfiguration resources

Create `runconfiguration_examples.yaml`:

```yaml
{{% readfile file="includes/dependent/resources/runconfiguration_examples.yaml" %}}
```

```bash
kubectl apply -f resources/runconfiguration_examples.yaml
```

Create `runconfiguration_training.yaml`:

```yaml
{{% readfile file="includes/dependent/resources/runconfiguration_training.yaml" %}}
```

Note that we have defined a `runConfigurations` trigger which creates a run of the training pipeline as soon as its dependency has produced a new model artifact.

```bash
kubectl apply -f resources/runconfiguration_training.yaml
```
