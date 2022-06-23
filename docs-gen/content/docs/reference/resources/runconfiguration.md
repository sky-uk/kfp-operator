---
title: "RunConfiguration"
weight: 2
---

The RunConfiguration resource represents the lifecycle of Recurring Runs (aka Jobs) on Kubeflow Pipelines.
Pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1
kind: RunConfiguration
metadata:
    name: penguin-pipeline-recurring-run
spec:
    pipelineName: penguin-pipeline
    experimentName: penguin-experiment
    schedule: '0 0 * * * *'
    runtimeParameters:
        TRAINING_RUNS: 100
```

## Fields

| Name | Description |
| --- | --- |
| `spec.pipelineName` | The name of the corresponding pipeline resource to run |
| `spec.experimentName` | The name of the corresponding experiment resource (Optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided) |
| `spec.schedule` | A cron schedule to execute training runs |
| `spec.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline |
