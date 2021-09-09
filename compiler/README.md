# KFP Components Compiler

This module compiles a TFX component defintion into a Kubeflow Pipelines file.

Requires a valid configuration file similar to [pipeline_conf.yaml](acceptance/pipeline_conf.yaml).
`spec.tfx_component` specified in this config file must be present on the python path.
`spec.env` variables get supplied to the components function at compilation time.

# Setup
```bash
pyenv install 3.7.10
pyenv local 3.7.10
poetry shell
poetry install
```

# Run the tests
```bash
poetry run pytest
```

# Build the compiler injector image
```bash
docker build . -t compiler
```