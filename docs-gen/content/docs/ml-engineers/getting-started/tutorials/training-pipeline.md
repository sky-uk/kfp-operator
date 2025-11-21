---
title: "Training Pipeline Tutorial"
linkTitle: "Training Pipeline"
description: "Complete tutorial for building and deploying TFX training pipelines with the KFP Operator"
weight: 10
---

# ML Pipeline Training Tutorial

This comprehensive tutorial walks you through creating, deploying, and managing a complete TFX training pipeline using the KFP Operator. You'll learn how to build a penguin species classification pipeline and manage its entire lifecycle through Kubernetes Custom Resources.

## What You'll Learn

By the end of this tutorial, you'll be able to:

- **Build TFX pipelines**
- **Containerize ML workflows** with proper dependencies
- **Deploy pipelines** using Kubernetes resources
- **Execute and monitor** pipeline runs
- **Set up automated scheduling** for continuous training
- **Handle events** for model deployment automation

## Prerequisites

Before starting, ensure you have:

- **KFP Operator installed** in your cluster ([Installation Guide](../../getting-started/installation/))
- **Docker** for building container images
- **Container registry access** (Docker Hub, GCR, ECR, etc.)
- **Basic TFX knowledge** (helpful but not required)

## Example Code

All code for this tutorial is available on [GitHub]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/master/quickstart).

```bash
# Clone the repository to follow along
git clone {{< param "github_repo" >}}.git
cd kfp-operator/docs-gen/includes/master/quickstart
```

## Step 1: Build the TFX Pipeline

We'll create a complete TFX pipeline for penguin species classification, based on the [TFX penguin example](https://www.tensorflow.org/tfx/tutorials/tfx/penguin_simple).

### Understanding the Pipeline Structure

Our pipeline consists of these TFX components:

1. **ExampleGen**: Ingests raw data from CSV files
2. **StatisticsGen**: Generates statistics for data analysis
3. **SchemaGen**: Infers data schema automatically
4. **ExampleValidator**: Validates data against schema
5. **Transform**: Performs feature engineering
6. **Trainer**: Trains the ML model
7. **Evaluator**: Evaluates model performance
8. **Pusher**: Deploys the trained model

### Create the Pipeline Definition

Create `penguin_pipeline/pipeline.py`:

{{% readfile file="/includes/master/quickstart/penguin_pipeline/pipeline.py" code="true" lang="python"%}}

### Create the Training Module

Create `penguin_pipeline/trainer.py`:

{{% readfile file="/includes/master/quickstart/penguin_pipeline/trainer.py" code="true" lang="python"%}}

### Create the Container Image

Create `Dockerfile`:

{{% readfile file="/includes/master/quickstart/Dockerfile" code="true" lang="dockerfile"%}}

### Build and Push the Container

Build the pipeline container and push to your registry:

```bash
# Set your container registry
export CONTAINER_REGISTRY="your-registry.com/your-project"

# Build the container
make docker-build

# Push to registry
make docker-push
```

### Verify the Build

Test your container locally:

```bash
# Run container to verify it works
docker run --rm ${REGISTRY}/penguin-pipeline:v1.0.0 python -c "
import tfx
print('Pipeline build successful!')
print(f'TFX version: {tfx.__version__}')
"
```

Expected output:
```
Pipeline build successful!
TFX version: 1.14.0
```

## 2. Create a Pipeline Resource

Now that we have a pipeline image, we can create a `pipeline.yaml` resource to manage the lifecycle of this pipeline on Kubeflow:

{{% readfile file="/includes/master/quickstart/resources/pipeline.yaml" code="true" lang="yaml"%}}

```bash
kubectl apply -f resources/pipeline.yaml
```

The pipeline now gets uploaded to Kubeflow in several steps. After a few seconds to minutes, the following command should result in a success:

```bash
kubectl get mlp

NAME               SYNCHRONIZATIONSTATE   PROVIDERID
penguin-pipeline   Succeeded              provider-namespace/provider-name
```

Now visit your Kubeflow Pipelines UI. You should be able to see the newly created pipeline named `penguin-pipeline`. 
Note that you will see two versions: 'penguin-pipeline' and 'v1'. This is due to an [open issue on Kubeflow](https://github.com/kubeflow/pipelines/issues/5881) where you can't specify a version when creating a pipeline.


## 3. Create a pipeline RunConfiguration resource

We can now define a recurring run declaratively using the `RunConfiguration` resource.

Note: remove `experimentName` if you want to use the `Default` experiment instead of `penguin-experiment`

Create `runconfiguration.yaml`:

{{% readfile file="/includes/master/quickstart/resources/runconfiguration.yaml" code="true" lang="yaml"%}}

```bash
kubectl apply -f resources/runconfiguration.yaml
```

This will trigger run of `penguin-pipeline` once every hour. Note that the cron schedule uses a 6-place space separated syntax as defined [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).

## 4. (Optional) Deploy newly trained models

If the operator has been installed with [Argo-Events](https://argoproj.github.io/argo-events/) support, we can now specify eventsources and sensors to update arbitrary Kubernetes config when a pipeline has been trained successfully.
In this example we are updating a serving component with the location of the newly trained model. 

Create `apply-model-location.yaml`. This creates an `EventSource` and a `Sensor` as well as an `EventBus`:

{{% readfile file="/includes/master/quickstart/resources/apply-model-location.yaml" code="true" lang="yaml"%}}

```bash
kubectl apply -f resources/apply-model-location.yaml
```
