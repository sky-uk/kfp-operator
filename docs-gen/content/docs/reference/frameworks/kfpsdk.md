---
title: "Kubeflow Pipelines SDK (KFP SDK)"
linkTitle: "KFP SDK"
type: docs
weight: 2
---

To create a KFP SDK pipeline:
1) Ensure your [Provider](../providers/overview/) supports KFP SDK by specifying the KFP SDK image in `spec.frameworks[]`.
2) Create a [Pipeline resource](../resources/pipeline/), specifying:
- the `kfpsdk` framework in `spec.framework.name`. This needs to match the name specified in the Provider.
- the path of the Python method that creates a list of TFX components under `spec.framework.parameters[].components`.
- any required [beam arguments](https://www.tensorflow.org/tfx/guide/beam#beam_pipeline_arguments) under `spec.framework.parameters[].beamArgs`.

## TFX Parameters

| Name       | Description                                                                                                                                                                                                                               |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline` | Fully qualified name of the Python function creating a KFP SDK pipeline. This function should be wrapped using the [`kfp.dsl.Pipeline` ](https://kubeflow-pipelines.readthedocs.io/en/2.0.0b6/source/dsl.html#kfp.dsl.pipeline)decorator. |

### KFP SDK Pipeline resource example
```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: kfpsdk-quickstart
spec:
  provider: provider-namespace/kfp
  image: kfp-operator-kfpsdk-quickstart:v1
  framework:
    name: kfpsdk
    parameters:
      pipeline: getting_started.pipeline.add_pipeline
```
