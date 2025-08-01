---
title: "Pipeline"
weight: 1
---

The Pipeline resource represents the lifecycle of ML pipelines.
Pipelines can be created, updated and deleted via this resource.
The operator compiles the pipeline into a deployable artifact while providing compile time parameters as environment variables.
It then submits the pipeline to Kubeflow and manages versions accordingly.

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Pipeline
metadata:
  name: penguin-pipeline
spec:
  provider: kfp
  image: kfp-quickstart:v1
  tfxComponents: base_pipeline.create_components
  env:
  - name: TRAINING_RUNS
    value: 100
```

## Fields

| Name                 | Description                                                                                             |
| -------------------- | ------------------------------------------------------------------------------------------------------- |
| `spec.image`         | Container image containing TFX component definitions.                                                   |
| `spec.tfxComponents` | Fully qualified name of the Python function creating pipeline components.                               |
| `spec.env[]`           | List of named objects. These will be provided to the `tfxComponents` function as environment variables. |
| `spec.beamArgs[]`      | List of named objects. These will be provided as `beam_pipeline_args` when compiling the pipeline.      |

## Versioning

Pipeline parameters can be updated at compile time. Pipeline versions therefore have to reflect both the pipelines image as well as its configuration. The operator calculates a hash over the pipeline spec and appends it to the image version to reflect this, for example: `v1-cf23df2207d99a74fbe169e3eba035e633b65d94`

## Identifier

A pipeline identifier field adheres to the following syntax:

`PIPELIE_NAME[:PIPELINE_VERSION]`
