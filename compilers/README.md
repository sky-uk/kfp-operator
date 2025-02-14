# KFP Components Compiler

This module compiles a TFX component defintion into a Kubeflow Pipelines file.

## Setup

Note that we use the [dynamic versioning plugin](https://pypi.org/project/poetry-dynamic-versioning/) for Poetry to version this module.
The version differs from those of the resulting containers (which are based on `git describe`) because Poetry would otherwise reject it. This will be installed automatically when running the `build` make target.

## Usage

This module is meant to be used inside Argo workflows as part of the Kubeflow Pipelines Operator.

It requires a valid configuration file similar to [pipeline_conf.yaml](tfx/acceptance/pipeline_conf.yaml).
`tfxComponents` specified in this config file must be present on the Python path.


```bash
PIPELINE_CONFIG=$(cat acceptance/pipeline_conf.yaml)
export PYTHONPATH="$PYTHONPATH:$(pwd)/../docs/quickstart/penguin_pipeline"
poetry run python kfp_compiler/compiler.py --pipeline_config "${PIPELINE_CONFIG}" --output_file out.yaml
```

## Run tests
```bash
make test
```

## Build the container image

The compiler injector image is only responsible for making the compiler available to a running container. It does NOT execute the compiler itself. This will be done by a workflow.

```bash
make docker-build
```