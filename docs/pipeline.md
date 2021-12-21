# Pipeline Resource

The Pipeline resource represents the lifecycle of TFX pipelines on Kubeflow Pipelines. Pipelines can be created, updated and deleted via this resource. The operator compiles the pipeline into a deployable artifact while providing compile time parameters as environment variables. It then submits the pipeline to kubeflow and manages versions accordingly.

```yaml
apiVersion: pipelines.kubeflow.com/v1
kind: Pipeline
metadata:
    name: penguin-pipeline
spec:
    image: kfp-quickstart:v1
    tfxComponents: base_pipeline.create_components
    env:
        TRAINING_RUNS: 100
```

## Fields

| Name | Description |
| --- | --- |
| `spec.image` | Container image containing TFX component definitions |
| `spec.tfxComponents` | Fully qualified name of the Python function creating pipeline components |
| `spec.env` | Dictionary of compile-time parameters. These will be provided to the `tfxComponents` function as environment variables |
| `spec.beamArgs` | Dictionary of Beam arguments. These will be provided as `beam_pipeline_args` when compiling the pipeline |

## Versioning

Pipeline parameters can be updated at compile time. Pipeline versions therefore have to reflect both the pipelines image as well as its configuration. The operator calculates a hash over the pipeline spec and appends it to the image version to reflect this on Kubeflow Pipelines, for example: `v1-cf23df2207d99a74fbe169e3eba035e633b65d94`
