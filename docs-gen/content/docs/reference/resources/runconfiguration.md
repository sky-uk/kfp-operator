---
title: "RunConfiguration"
weight: 2
---

The RunConfiguration resource represents the lifecycle of recurring runs (aka Jobs in KFP).
Pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: RunConfiguration
metadata:
  name: penguin-pipeline-recurring-run
spec:
  run:
    pipeline: penguin-pipeline:v1-abcdef
    experimentName: penguin-experiment
    runtimeParameters:
    - name: TRAINING_RUNS
      value: '100'
  triggers:
  - schedule:
      cronExpression: '0 * * * *'
  - onChange: {}
```

A Run Configuration can have one of more triggers that determine when the next training run will be started.

## Fields

| Name                        | Description                                                                                                                                                                                                                                       |
|-----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.run.pipeline`         | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will track the latest version of the specified pipeline.                                           |
| `spec.run.experimentName`   | The name of the corresponding experiment resource (optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided). |
| `spec.run.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline.                                                                                                                                                                                 |
| `spec.triggers[]`       | Describe the kind of event that will start a run.                                                                                                                                                                                                 |

Each trigger type can accept other type-specific parameters.

### Scheduled Trigger

Runs are executed on a schedule.

| Parameter        | Description                                                                                                                                                                                        |
|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `cronExpression` | A cron schedule to execute training runs. It can have 5 (standard cron) or 6 (first digit expresses seconds) fields. When a provider does not support the 6-field format, seconds will be omitted. |

### On-Change Trigger

Runs are executed when the referenced pipeline changes.
