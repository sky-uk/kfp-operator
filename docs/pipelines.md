## Pipeline Resource

The pipeline resource represents the lifecycle of TFX pipelines on Kubeflow. Pipelines can be create, updated and deleted via this resource. The operator compiles the pipeline into a deployable artifact while providing compile time parameters as environment variables. It the submits the pipeline to kubeflow and manages versions accordingly.

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

### Fields

| Name | Description |
| --- | --- |
| `spec.image` |docker image containing TFX component definitions|
| `spec.tfxComponents` | fully qualified name of the python function creating pipeline components |
| `spec.env` | dictionary of compile-time parameters. These will be provided to the `tfxComponents` function as environment variables |

### Versioning

Pipeline parameters can be updated at compile time. Pipeline versions therefore have to reflect both the pipelines image as well as its configuration. The operator calculates a hash over the pipeline spec and appends it to the image version to reflect this on Kubeflow Pipelines. The pipeline above would result in the following version: `v1-TODO`