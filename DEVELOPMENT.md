# Development

<!-- TOC -->
* [Development](#development)
  * [Set up the development environment](#set-up-the-development-environment)
  * [Run unit tests](#run-unit-tests)
  * [Building and Publishing](#building-and-publishing)
    * [Building and publishing container images](#building-and-publishing-container-images)
    * [Building and publishing the Helm chart](#building-and-publishing-the-helm-chart)
  * [Running locally](#running-locally)
    * [Getting started](#getting-started)
    * [Rebuild after code changes](#rebuild-after-code-changes)
    * [Tear down](#tear-down)
    * [How the stubs work](#how-the-stubs-work)
  * [Run Argo integration tests](#run-argo-integration-tests)
  * [Coding Guidelines](#coding-guidelines)
    * [Logging](#logging)
  * [CRD Versioning](#crd-versioning)
<!-- TOC -->

We use [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to scaffold the kubernetes controllers.
The [Kubebuilder Book](https://book.kubebuilder.io/) is a good introduction to the topic and we recommend reading it before proceeding.

Please install kubebuilder `3.15.1` before upgrading or adding new Custom Resources:
```
asdf plugin-add kubebuilder https://github.com/virtualstaticvoid/asdf-kubebuilder.git
asdf install kubebuilder 3.15.1
```
Note that we currently need to use kubebuilder version 3 or below, as [work is required to support later versions](https://github.com/sky-uk/kfp-operator/issues/381).

## Set up the development environment

Install Go by following the instructions on the [website](https://golang.org/doc/install).

We use [asdf](http://asdf-vm.com) to set up the development environment. Install it following the [Getting Started Guide](http://asdf-vm.com/guide/getting-started.html).

Add the required asdf plugins:

```sh
asdf plugin add python
asdf plugin add protoc
asdf plugin add hugo
asdf plugin add golang
asdf plugin add nodejs
asdf plugin add minikube
asdf plugin add uv
```

Install all tool versions as follows:

```sh
asdf install
```

Many commands in this guide will run *against your current kubernetes context*; make sure that it is set accordingly. [Minikube](https://minikube.sigs.k8s.io/docs/start/) provides a local Kubernetes cluster ideal for development.

## Run unit tests

```sh
make unit-test
```

Note: on first execution, the test environment will get downloaded and the command will therefore take longer to complete.

## Building and Publishing

### Building and publishing container images

Build the operator's container images as follows:

```sh
make docker-build docker-build-argo
```

Push to the container registry:

```sh
CONTAINER_REPOSITORIES=<YOUR_CONTAINER_REPOSITORY> make docker-push
```

For example, to push to Google Artifact Registry:

```sh
CONTAINER_REPOSITORIES=europe-docker.pkg.dev/<PROJECT_NAME>/images make docker-push
```

### Building and publishing the Helm chart

Build the Helm chart as follows:

```sh
make helm-package
```

Push the Helm chart using one of the following options:

1. OCI Image

```sh
HELM_REPOSITORIES=oci://<YOUR_CHART_REPOSITORY> make helm-publish
```

For example, to push to Google Artifact Registry:

```sh
HELM_REPOSITORIES=oci://europe-docker.pkg.dev/<PROJECT_NAME>/charts make helm-publish
```

2. `.tar.gz` archive

```sh
HELM_REPOSITORIES=https://<YOUR_CHART_REPOSITORY> NETRC_FILE=<YOUR_NETRC_FILE> make helm-publish
```

Provide an optional [.netrc file](https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html) for credentials:

## Running locally

The local development environment uses a **stub provider** so you can test the full operator reconciliation loop without any real training infrastructure (no Vertex AI, no Kubeflow Pipelines, etc.).

### Getting started

```sh
# Spin everything up (minikube, argo, cert-manager, minio, operator, eventing, stub provider)
make minikube-up

# Apply resources
kubectl apply -f local/pipeline.yaml
kubectl apply -f local/runconfiguration.yaml

# Watch them reconcile
kubectl get pipelines -w
kubectl get runconfigurations -w
kubectl get runschedules -w
kubectl get runs -w
```

After a minute or so, `stub-pipeline` should reach `Succeeded` and `stub-runconfiguration` should be actively creating RunSchedules and a Run.
The Run should reach `Succeeded` and then be deleted. Shortly after the Run is created, the stub provider fires a fake run completion event.
You can verify the full eventing flow by checking the Sensor logs — it should log `Successfully processed trigger 'log'` once the event has passed through the webhook, trigger service, NATS, and EventSource.

### Rebuild after code changes

```sh
make minikube-upgrade
```

This rebuilds the operator, stub provider, and stub compiler images, pushes them to the in-cluster registry, and does a `helm upgrade`.

### Tear down

```sh
make minikube-down
```

### How the stubs work

- The stub provider (`provider-service/stub`) implements the full provider gRPC interface but doesn't talk to any external service.
  Every operation (create/update/delete for pipelines, runs, run schedules, and experiments) returns a hardcoded success response immediately.
  This means the operator's reconciliation loop runs end-to-end — creating Argo Workflows, waiting for them to complete, updating status — without needing real infrastructure.
  - When a run is created, the stub provider also fires a fake run completion event to the operator's webhook after a 5-second delay (simulating pipeline execution time).
    This exercises the full eventing flow: the event is received by the webhook, forwarded via gRPC to the run completion event trigger, published to NATS, picked up by the Argo Events EventSource, and delivered to the Sensor.

- The stub compiler (`compilers/stub`) is equally minimal. Its `compile.sh` writes a static `{"foo": "bar"}` JSON file as the compiled pipeline output, and its `inject.sh` simply copies the compile script into the workflow step.
  This is enough for the operator to treat the pipeline as successfully compiled.

- The stub Provider CR (`local/stub-provider.yaml`) registers both `stub` and `tfx` as supported frameworks, both pointing at the stub compiler image.
  This lets you test with either framework name in your Pipeline resources.

## Run Argo integration tests

To run integration tests, we currently require a one-off setup of the Kubernetes cluster:

```sh
make integration-test-up
```

You can now run the integration tests as follows (this can take several minutes to complete):

```sh
make integration-test
```

Finally, bring down the environment after your tests:

```sh
make integration-test-down
```

## Coding Guidelines

### Logging

We use the [zap](https://github.com/uber-go/zap) implementation of [logr](https://github.com/go-logr/logr) via the [zapr](https://github.com/go-logr/zapr) module.

[Verbosity levels](https://github.com/go-logr/logr#why-v-levels) are set according to the following rules:

| Zap Level          | Description                                               | Example                                                    |
| ------------------ | --------------------------------------------------------- | ---------------------------------------------------------- |
| 0, `error`, `info` | Will always be logged. Appropriate for all major actions. | state transitions, errors                                  |
| 1, `debug`         | Appropriate for high-level technical information.         | resource creation/update/deletion                          |
| 2                  | Appropriate for low-level technical information.          | resource retrieval, finalizers, profiling, expected errors |
| 3                  | Appropriate for verbose debug statements.                 | printing resources                                         |

## CRD Versioning

Steps to create a new CRD version:

1) Copy and paste the current `hub` directory into its version named directory, e.g. `v1alpha1`.
1) Change `groupversion_info.go` to reflect the new version, e.g. `v1alpha2`.
1) Change all the package names in the `hub` directory to be the new required version, e.g. `v1alpha2` (but leave the directory called `hub`, all Client code imports hub directory and names it, so that no changes needed on creating new version).
1) Change all the conversion functions in the now old version (e.g. `v1alpha1`) to convert to and from the new hub version (e.g. `v1alpha2`).
1) Make the required schema changes in `hub`. This will involve changing the conversion code in all older versions.
1) Register the old schema (e.g. `v1alpha1`) in the controller runtime (see [here](main.go#L56)).
1) Copy changes into helm version of the CRD to match that generated in `config/crd/bases`.
1) Set the new version as the default stored version in `helm/kfp-operator/values.yaml`, e.g. `manager.multiversion.storedVersion: v1alpha2`
