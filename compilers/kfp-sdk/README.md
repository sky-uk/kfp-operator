### KFP SDK Compiler

KFP SDK Compiler is a tool that compiles a KFP SDK pipeline definition into a Kubeflow Pipelines representation.

### Usage

KFP SDK compiler is compatible with KFP SDK [2.12.2](https://kubeflow-pipelines.readthedocs.io/en/sdk-2.12.1/), supports Python 3.9 onwards and requires minimum of `v1beta1` Pipeline resource definition with:
- framework `type` set to `kfpsdk`
- `pipeline` parameter set to fully qualified name of Python function decorated with `@dsl.pipeline`, separated by `.`

```yaml
---
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: quickstart
spec:
  provider: vai
  image: kfpsdk-quickstart:v1
  framework:
    type: kfpsdk
    parameters:
       pipeline: quickstart.pipeline_function
```

### Sample KFP SDK pipeline

> [!IMPORTANT]
> Setting @dsl.pipeline `name` and `pipeline_root` is not currently supported and will be overwritten by the compiler.

```python
from kfp import dsl

@dsl.component
def component():
    pass

@dsl.pipeline(
    description='A simple pipeline to get started with the KFP SDK.'
)
def pipeline_function():
    component()
```

### Compiler parameters
The compiler requires the following mandatory flags/parameters:

- `--pipeline_config`: Path to the pipeline configuration file in YAML format.
- `--output_file`: Path where the compiled output will be saved.
- `--provider_config`: Path to the provider configuration file in YAML/json format.
