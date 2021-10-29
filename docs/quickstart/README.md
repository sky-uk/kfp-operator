# Quickstart

This tutorial walks you through the creation of a simple TFX pipeline and shows you how to manage pipelines via Kubernetes Custom Resources.

We assume that you are already familiar with TFX and Kubeflow and that you have access to a running instance of Kubeflow Pipelines.

## Build the Pipeline

We use the same pipeline as the [TFX example](https://www.tensorflow.org/tfx/tutorials/tfx/penguin_simple) with a few modifications. Note that the pipeline definition itself is simpler because all infrastructure refences, like pusher and pipeline root, will be injected by the operator before the pipeline is uploaded to Kubeflow.

```bash
docker build . -t kfp-quickstart:v1
...
docker push kfp-quickstart:v1
```

## Create a Pipeline Resource

Now that we have a pipeline image, we can create a `pipeline` resource to manage the lifecycle of this pipeline on Kubeflow:

```bash
cat << EOF | kubectl apply -f
apiVersion: pipelines.kubeflow.org/v1
kind: Pipeline
metadata:
    name: penguin-pipeline
    spec:
        image: kfp-quickstart:v1
EOF
```

The pipeline now gets uploaded to Kubeflow in several steps. After a few seconds to minutes, the following command should result in a success:

```bash
kubectl get pipeline

NAME               SYNCHRONIZATIONSTATE   KFPID
penguin-pipeline   Succeeded              53905abe-0337-48de-875d-67b9285f3cf7
```

Now visit you Kubeflow Pipelines UI. You should be able to see the newly created pipeline named `penguin-pipeline`. Note that you will see two versions: 'penguin-pipeline' and 'v1'. This is due to an [open issue on Kubeflow](https://github.com/kubeflow/pipelines/issues/5881) where you can't specify a version when creating a pipeline.

## Create a pipeline RunConfiguration resource

We can now define a recurring run delcaratively using the `RunConfiguration` resource:

```bash
cat << EOF | kubectl apply -f
apiVersion: pipelines.kubeflow.com/v1
kind: RunConfiguration
metadata:
    name: penguin-pipeline-recurring-run
    spec:
        pipelineName: penguin-pipeline
        schedule: '0 0 * * * *'
EOF
```

This will trigger run of `penguin-pipeline` once every hour. Note that the cron schedule uses a 6-place space separated syntax as defined [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).

## Update the serving model location when the pipeline run succeeds

Using [Argo Event](https://argoproj.github.io/argo-events/) we can specify arbitrary actions on pipeline run completions.
In this example, we apply a `ConfigMap` containing the model's serving location when the pipeline run succeeds.

```bash
kubectl apply -f apply-model-location.yaml
```

This creates Argo Events `EventSource` and `Sensor` which listen to pipeline run updates and apply the following configmap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
    name: serving-config
data:
    serving-model: {{serving-model-location}}
```

