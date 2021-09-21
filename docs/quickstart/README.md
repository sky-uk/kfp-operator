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

NAME               STATUS             PIPELINEID
penguin-pipeline   Succeeded          53905abe-0337-48de-875d-67b9285f3cf7
```

Now visit you Kubeflow Pipelines UI. You should be able to see the newly created pipeline named `penguin-pipeline`. Note that you will see two versions: 'penguin-pipeline' and 'v1'. This is due to an [open issue on Kubeflow](https://github.com/kubeflow/pipelines/issues/5881) where you can't specify a version when creating a pipeline.
