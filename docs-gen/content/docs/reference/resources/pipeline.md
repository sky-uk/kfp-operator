---
title: "Pipeline"
weight: 1
---

The Pipeline resource represents the lifecycle of ML pipelines.
Pipelines can be created, updated and deleted via this resource.
The operator compiles the pipeline into a deployable artifact while providing compile time parameters as environment
variables.
It then submits the pipeline to Kubeflow and manages versions accordingly.

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
  env:
  - name: TRAINING_RUNS
    value: 100
```

## Fields

| Name                        | Description                                                                                                                                                                 |
|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.provider`             | The namespace and name of the associated [Provider resource](../provider/) separated by a `/`, e.g. `provider-namespace/provider-name`.                                     |
| `spec.image`                | Container image containing TFX component definitions.                                                                                                                       |
| `spec.env[]`                | List of named objects. These will be provided by the compiler to the pipeline/components function as environment variables                                                  |
| `spec.framework.name`       | Sets a specific [pipeline framework](../../ml-engineers/frameworks) to use.                                                                                                 |
| `spec.framework.parameters` | Parameters to pass to the pipeline framework compiler. A map of any parameters required by that framework can be passed, e.g. `components: base_pipeline.create_components` |

## Versioning

Pipeline parameters can be updated at compile time. Pipeline versions therefore have to reflect both the pipelines image
and its configuration. The operator calculates a hash over the pipeline spec and appends it to the image version
to reflect this, for example: `v1-cf23df2207d99a74fbe169e3eba035e633b65d94`

## Identifier

A pipeline identifier field adheres to the following syntax:

`PIPELIE_NAME[:PIPELINE_VERSION]`
