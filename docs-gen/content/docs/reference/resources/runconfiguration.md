---
title: "RunConfiguration"
weight: 2
---

The RunConfiguration resource represents the lifecycle of Recurring Runs (aka Jobs) on Kubeflow Pipelines.
Pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha2
kind: RunConfiguration
metadata:
    name: penguin-pipeline-recurring-run
spec:
    pipeline: penguin-pipeline:v1-abcdef
    experimentName: penguin-experiment
    schedule: '0 0 * * * *'
    runtimeParameters:
    - name: TRAINING_RUNS
      value: 100
```

## Fields

| Name | Description |
| --- | --- |
| `spec.pipeline` | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will track the latest version of the specified pipeline. |
| `spec.experimentName` | The name of the corresponding experiment resource (Optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided) |
| `spec.schedule` | A cron schedule to execute training runs |
| `spec.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline |
