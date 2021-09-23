# KFP Components Compiler

This module compiles a TFX component defintion into a Kubeflow Pipelines file.

## Setup
```bash
pyenv install -s
poetry install
```

## Usage

This module is meant to be used inside Argo workflows as part of the Kubeflow Pipelines Operator.

It requires a valid configuration file similar to [pipeline_conf.yaml](acceptance/pipeline_conf.yaml).
`spec.tfxComponents` specified in this config file must be present on the Python path.


```bash
PIPELINE_CONFIG=$(cat acceptance/pipeline_conf.yaml)
export PYTHONPATH="$PYTHONPATH:$(pwd)/../docs/quickstart/penguin_pipeline"
poetry run python compiler/compiler.py --pipeline_config "${PIPELINE_CONFIG}" --output_file out.yaml
```

## Run tests
```bash
poetry run pytest
```

## Build the container image

The compiler injector image is only responsible for making the compiler available to a running container. It does NOT execute the compiler itself. This will be done by a workflow.

```bash
docker build . -t compiler
```