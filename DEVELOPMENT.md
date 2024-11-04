# Development

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

We use [asdf](http://asdf-vm.com) to set up the development environment. Install it it following the [Getting Started Guide](http://asdf-vm.com/guide/getting-started.html).
Install all tool versions as follows:

```bash
asdf install
```

Many commands in this guide will run *against your current kubernetes context*; make sure that it is set accordingly. [Minikube](https://minikube.sigs.k8s.io/docs/start/) provides a local Kubernetes cluster ideal for development.

## Run unit tests

```sh
make test
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
CONTAINER_REPOSITORIES=<YOUR_CONTAINER_REPOSITORY> make docker-push docker-push-argo
```

For example, to push to Google Artifact Registry:

```sh
CONTAINER_REPOSITORIES=europe-docker.pkg.dev/<PROJECT_NAME>/images make docker-push docker-push-argo
```

### Building and publishing the Helm chart

Build the Helm chart as follows:

```shell
make helm-package
```

Push the Helm chart using one of the following options:

1. OCI Image
```shell
HELM_REPOSITORIES=oci://<YOUR_CHART_REPOSITORY> make helm-publish
```

For example, to push to Google Artifact Registry::

HELM_REPOSITORIES=oci://europe-docker.pkg.dev/<PROJECT_NAME>/charts make helm-publish

2. `.tar.gz` archive

```shell
HELM_REPOSITORIES=https://<YOUR_CHART_REPOSITORY> NETRC_FILE=<YOUR_NETRC_FILE> make helm-publish
```

Provide an optional [.netrc file](https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html) for credentials:

## Running locally

Configure the controller to your environment in [controller_manager_config.yaml](../../config/manager/controller_manager_config.yaml) replacing the placeholders (see [docs](../README.md#configuration)).

Next install Custom Resource Definitions and run the controller:

```sh
make install
make run
```

CRDs will be installed into an existing Kubernetes cluster. A running instance of Kubeflow is required on that cluster. The controller will run locally, interacting with the remote Kubernetes API.

Please refer to the [quickstart tutorial](../quickstart) for instructions on creating a sample pipeline resource.

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

| Zap Level | Description | Example |
| --- | --- | --- |
| 0, `error`, `info` | Will always be logged. Appropriate for all major actions. | state transitions, errors |
| 1, `debug` | Appropriate for high-level technical information. | resource creation/update/deletion |
| 2 | Appropriate for low-level technical information. | resource retrieval, finalizers, profiling, expected errors |
| 3 | Appropriate for verbose debug statements. | printing resources |
