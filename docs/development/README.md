# Contributing and Development

We use [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to scaffold the kubernetes controllers.
The [Kubebuilder Book](https://book.kubebuilder.io/) is a good introduction to the topic and we recommend reading it before proceeding.

## Set up the development environment

Install Go by following the instructions on the [website](https://golang.org/doc/install).

Many commands in this guide will run *against your current kubernetes context*; make sure that it is set accordingly. [Minikube](https://minikube.sigs.k8s.io/docs/start/) provides a local Kubernetes cluster ideal for development.

## Run unit tests

```sh
make test
```

Note: on first execution, the test environment will get downloaded and the command will therefore take longer to complete.

## Running locally

Build all images as follows:

```sh
(cd compiler; docker build . -t ${YOUR_REGISTRY}/compiler; docker push ${YOUR_REGISTRY}/compiler)
(cd kfp-tools; docker build . -t ${YOUR_REGISTRY}/kfp-tools; docker push ${YOUR_REGISTRY}/kfp-tools)
```

Configure the controller to your environment in [controller_manager_config.yaml](../../config/manager/controller_manager_config.yaml)

Next install Custom Resource Defitions and run the controller:

```sh
make install
make run
```

CRDs will be installed into an existing Kubernetes cluster. A running instance of Kubeflow is required on that cluster. The controller will run locally, interacting with the remote Kubernetes API.

Please refer to the [Quickstart tutorial](../quickstart) for instructions on creating a sample pipeline resource.

## Run Argo integration tests

To run integration tests, we currently require a one-off setup of the Kubernetes cluster:

```sh
# Install Argo
kubectl create ns argo
kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo-workflows/master/manifests/quick-start-postgres.yaml
# Start wiremock
kubectl apply -f integration_tests/wiremock.yaml
kubectl port-forward service/kfp-wiremock 8081:80
# Proxy the API server
kubectl proxy --port=8080
```

Next, build all relevant images in the minikube docker environment.
Note: alternatively, you can build the images locally and make them available using `minikube image load`.

```sh
eval $(minikube -p argo-integration-tests docker-env)
(cd docs/quickstart; docker build . -t kfp-quickstart)
(cd compiler; docker build . -t compiler)
(cd kfp-tools; docker build . -t kfp-tools)
```

You can now run the integration tests as follows:
```sh
make integration-test
```
