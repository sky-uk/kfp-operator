# Contributing and Development

We use [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to scaffold the kubernetes controllers.
The [Kubebuilder Book](https://book.kubebuilder.io/) is a good introduction to the topic and we recommend reading it before proceeding.

## Set up the development environment

Install go by following the [website](https://golang.org/doc/install)

Install the dependencies:

```sh
go get
```

## Running locally

The following command wil run the controller locally *against your current kubernetes context*.
This means that CRDs will be installed into an existing k8s cluster, but the controller will run locally, interacting with the rempote k8s API.

```sh
make install
make run
```

## Run the tests

Note: on first execution, the test environment will get downloaded and the command will therefore take longer to complete.

```sh
make test
```

## Run argo integration tests

```sh
# Start minikube
minikube start -p argo-integration-tests --driver=hyperkit
# Install argo
kubectl create ns argo
kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo-workflows/master/manifests/quick-start-postgres.yaml
# proxy the minikube API server
kubectl proxy --port=8080
# Start wiremock
kubectl apply -f controllers/integration_tests/wiremock.yaml
kubectl port-forward -n argo $(kubectl get pod -n argo --selector="app=kfp-wiremock" --output jsonpath='{.items[0].metadata.name}') 8081:8080
# Load all images into minikube
minikube -p argo-integration-tests image load kfp-tools
minikube -p argo-integration-tests image load compiler
minikube -p argo-integration-tests image load test-pipeline

# Run the tests
integration-test
```

