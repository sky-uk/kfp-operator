# RunConfiguration Resource

The RunConfiguration resource represents the lifecycle of Recurring Runs (aka Jobs) on Kubeflow Pipelines.
Pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1
kind: RunConfiguration
metadata:
    name: penguin-pipeline-recurring-run
spec:
    pipelineName: penguin-pipeline
    schedule: '0 0 * * * *'
    runtimeParameters:
        TRAINING_RUNS: 100
```

Note: The experiment will be created if it does not exist when creating the run configuration.

## Fields

| Name | Description |
| --- | --- |
| `spec.pipelineName` | The name of the corresponding pipeline resource to run |
| `spec.schedule` | A cron schedule to execute training runs |
| `spec.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline |
