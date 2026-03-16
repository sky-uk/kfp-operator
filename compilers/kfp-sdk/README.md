### KFP SDK Compiler

KFP SDK Compiler is a tool that compiles a KFP SDK pipeline definition into a Kubeflow Pipelines representation.

### Usage

KFP SDK compiler is compatible with KFP SDK [2.12.2](https://kubeflow-pipelines.readthedocs.io/en/sdk-2.12.1/), supports Python 3.9 onwards and requires minimum of `v1beta1` Pipeline resource definition with:
- `spec.framework.type` set to `kfpsdk`
- `spec.framework.parameters[].pipeline` set to fully qualified name of Python function decorated with `@dsl.pipeline`, delimited by `.`

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

### Environment Variables

The compiler automatically injects the following environment variables during compilation:

- `KFP_PIPELINE_IMAGE`: Set to the value of `spec.image` from the Pipeline resource. This can be used in pipeline code to dynamically set component base images.

### Sample KFP SDK pipeline

> [!IMPORTANT]
> Setting @dsl.pipeline `name` and `pipeline_root` is not currently supported and will be overwritten by the compiler.

```python
import os
from kfp import dsl

# Use the image specified in the Pipeline resource
DEFAULT_IMAGE = os.environ.get("KFP_PIPELINE_IMAGE", "python:3.9")

@dsl.component(base_image=DEFAULT_IMAGE)
def component():
    pass

@dsl.pipeline(
    description='A simple pipeline to get started with the KFP SDK.'
)
def pipeline_function():
    component()
```
