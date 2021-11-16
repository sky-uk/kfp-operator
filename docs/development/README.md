# Contributing and Development

We use [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to scaffold the kubernetes controllers.
The [Kubebuilder Book](https://book.kubebuilder.io/) is a good introduction to the topic and we recommend reading it before proceeding.

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

## Running locally

Build all images as follows:

```sh
make docker-build-argo
```

Push to the container registry used by the Kubernetes cluster:

```sh
export CONTAINER_REGISTRY_HOSTS=host:port # <- replace this
make docker-push-argo
```

Configure the controller to your environment in [controller_manager_config.yaml](../../config/manager/controller_manager_config.yaml) replacing the placeholders (see [docs](../README.md#configuration)).

Next install Custom Resource Defitions and run the controller:

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

You can now run the integration tests as follows:
```sh
make integration-test
```

Finally, bring down the environment after your tests:

```sh
make integration-test-down
```

## Coding Guidelines

### Logging

Log [verbosity levels](https://github.com/go-logr/logr#why-v-levels) should be set according to the following rules:

| Level | Description | Example |
| --- | --- | --- |
| 0 | Will always be logged. Appropriate for all major actions. | state transitions, errors |
| 1 | Appropriate for high-level technical information. | resource creation/update/deletion |
| 2 | Appropriate for low-level technical information. | resource retrieval, finalizers, profiling, expected errors |
