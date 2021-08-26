# Contributing and Development

We use [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to scaffold the kubernetes controllers.
The [Kubebuilder Book](https://book.kubebuilder.io/) is a good introduction to the topic and we recommend reading it before proceeding.

## Set up the development environment

Install go by following the [website](https://golang.org/doc/install)

Install the dependencies:

```sh
go get
go mod vendor
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

Using cached envtest tools from /Users/jmd31/projects/kfp-operator/testbin
setting up env vars
?       github.com/sky-uk/kfp-operator  [no test files]
ok      github.com/sky-uk/kfp-operator/api/v1   0.484s  coverage: 22.0% of statements
ok      github.com/sky-uk/kfp-operator/controllers      12.761s coverage: 83.3% of statements
?       github.com/sky-uk/kfp-operator/external [no test files]
```