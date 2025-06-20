---
title: "TensorFlow Extended (TFX)"
linkTitle: "TFX"
type: docs
weight: 1
---

To create a TFX pipeline:
1) Ensure your [Provider](../providers/overview/) supports TFX by specifying the TFX image in `spec.frameworks[]`.
2) Create a [Pipeline resource](../resources/pipeline/), specifying:
- the `tfx` framework in `spec.framework.name`. This needs to match the name specified in the Provider.
- the path of the Python method that creates a list of TFX components under `spec.framework.parameters[].components`.
- any required [beam arguments](https://www.tensorflow.org/tfx/guide/beam#beam_pipeline_arguments) under `spec.framework.parameters[].beamArgs`.


## TFX Parameters

| Name         | Description                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------------- |
| `components` | Fully qualified name of the Python function creating pipeline components.                          |
| `beamArgs[]` | List of named objects. These will be provided as `beam_pipeline_args` when compiling the pipeline. |
|              |


### TFX Pipeline resource example
```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: penguin-pipeline
spec:
  provider: provider-namespace/provider-name
  image: quickstart:v1
  framework:
    name: tfx
    parameters:
      components: base_pipeline.create_components
      beamArgs:
      - name: anArg
        value: aValue
  env:
  - name: TRAINING_RUNS
    value: 100
```
