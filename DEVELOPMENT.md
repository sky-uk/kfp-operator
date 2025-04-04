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
CONTAINER_REPOSITORIES=<YOUR_CONTAINER_REPOSITORY> make docker-push
```

For example, to push to Google Artifact Registry:

```sh
CONTAINER_REPOSITORIES=europe-docker.pkg.dev/<PROJECT_NAME>/images make docker-push
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

Running `make minikube-up` will spin up a local K8s cluster with the operator deployed.

You can optionally perform additional provider setup and teardown steps by including a `provider-setup.sh` and `provider-teardown.sh` script.

CRDs will be installed into a local [minikube](https://github.com/kubernetes/minikube) cluster.

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

| Zap Level          | Description                                               | Example                                                    |
| ------------------ | --------------------------------------------------------- | ---------------------------------------------------------- |
| 0, `error`, `info` | Will always be logged. Appropriate for all major actions. | state transitions, errors                                  |
| 1, `debug`         | Appropriate for high-level technical information.         | resource creation/update/deletion                          |
| 2                  | Appropriate for low-level technical information.          | resource retrieval, finalizers, profiling, expected errors |
| 3                  | Appropriate for verbose debug statements.                 | printing resources                                         |

## CRD Versioning

Steps to create a new CRD version:

1) Copy and paste the current `hub` into its version named directory, e.g. `v1alpha1`.
1) Change groupversion_info.go to reflect the new version, e.g. `v1alpha2`.
1) Change all the package names in the hub directory to be the new required version, e.g. `v1alpha2` (but leave the directory called `hub`, all Client code imports hub directory and names it, so that no changes needed on creating new version).
1) Change all the conversion functions in the now old version (e.g. `v1alpha1`) to convert to and from the new hub version (e.g. `v1alpha2`).
1) Make the required schema changes in `hub`. This will involve changing the conversion code in all older versions.
1) Register the old schema (e.g. `v1alpha1`) in the controller runtime (see [here](main.go#L56)).
1) Copy changes into helm version of the CRD to match that generated in `config/crd/bases`.

Steps to install a new CRD version:

1) Make sure `.Values.manager.storedVersion` in the helm values is set to the currently stable version. This ensures that the new version will not become the stored version until all changes for that version have been released.
1) Install/upgrade the operator in the desired cluster as normal.
    - Note that even though the new version will not be stored, K8s will still be able to serve it and clients like `kubectl` will, in fact, get this new version by default (in accordance with the [version priority algorithm](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#version-priority)). This means that when users retrieve their resources, K8s will convert them to the new version, even if they were created with the old one. This shouldn't cause any problems as long as the conversion webhook logic has been implemented correctly. 
    - Setting `served: false` on the new version to hide it from users until it's ready to be used is not an option because (if you followed the steps above when creating the new version in the code) the operator will always try to get the hub version, and if K8s can't serve that version then the controller manager will error and crash.
1) The operator can continue to be incrementally upgraded with changes that are to be bundled in the new version.

Steps to roll back a CRD version _that hasn't been stored_:

1) If a new version hasn't been set to the stored version yet, then a rollback of any commit since the introduction of the new version (including the commit that added the version in the first place) can be performed simply by reverting the offending commit and releasing the operator again.

