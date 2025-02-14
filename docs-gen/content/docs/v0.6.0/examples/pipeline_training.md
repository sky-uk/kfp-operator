---
title: "Pipeline Training"
weight: 1
---

This tutorial walks you through the creation of a simple TFX pipeline on Kubeflow Pipelines and shows you how to manage pipelines via Kubernetes Custom Resources.

The examples for this tutorial can be found on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/quickstart).

## 1. Build the Pipeline

We use the same pipeline as the [TFX example](https://www.tensorflow.org/tfx/tutorials/tfx/penguin_simple) with a few modifications.

Create `pipeline.py`.
Note that the pipeline definition itself is simpler because all infrastructure references, like pusher and pipeline root, will be injected by the operator before the pipeline is uploaded to Kubeflow:

{{% readfile file="/includes/quickstart/penguin_pipeline/pipeline.py" code="true" lang="python"%}}

Create `trainer.py`.
The training code remains unchanged:

{{% readfile file="/includes/quickstart/penguin_pipeline/trainer.py" code="true" lang="python"%}}

Create `Dockerfile`.

{{% readfile file="/includes/quickstart/Dockerfile" code="true" lang="dockerfile"%}}

Next, build the pipeline as a Docker container and push it:

```bash
docker build . -t kfp-quickstart:v1
...
docker push kfp-quickstart:v1
```

## 2. Create a Pipeline Resource

Now that we have a pipeline image, we can create a `pipeline.yaml` resource to manage the lifecycle of this pipeline on Kubeflow:

{{% readfile file="/includes/quickstart/resources/pipeline.yaml" code="true" lang="yaml"%}}

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

{{% readfile file="/includes/quickstart/resources/experiment.yaml" code="true" lang="yaml" %}}

```bash
kubectl apply -f resources/experiment.yaml
```

## 4. Create a pipeline RunConfiguration resource

We can now define a recurring run declaratively using the `RunConfiguration` resource.

Note: remove `experimentName` if you want to use the `Default` experiment instead of `penguin-experiment`

Create `runconfiguration.yaml`:

{{% readfile file="/includes/quickstart/resources/runconfiguration.yaml" code="true" lang="yaml"%}}

```bash
kubectl apply -f resources/runconfiguration.yaml
```

This will trigger run of `penguin-pipeline` once every hour. Note that the cron schedule uses a 6-place space separated syntax as defined [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).

## 5. (Optional) Deploy newly trained models

If the operator has been installed with [Argo-Events](https://argoproj.github.io/argo-events/) support, we can now specify eventsources and sensors to update arbitrary Kubernetes config when a pipeline has been trained successfully.
In this example we are updating a serving component with the location of the newly trained model. 

Create `apply-model-location.yaml`. This creates an `EventSource` and a `Sensor` as well as an `EventBus`:

{{% readfile file="/includes/quickstart/resources/apply-model-location.yaml" code="true" lang="yaml"%}}

```bash
kubectl apply -f resources/apply-model-location.yaml
```
