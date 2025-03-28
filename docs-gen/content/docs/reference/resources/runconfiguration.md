---
title: "RunConfiguration"
weight: 2
---

The RunConfiguration resource represents the lifecycle of recurring runs (aka Jobs in KFP).
Pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: RunConfiguration
metadata:
  name: penguin-pipeline-recurring-run
spec:
  run:
    provider: provider-namespace/kfp
    pipeline: penguin-pipeline:v1-abcdef
    experimentName: penguin-experiment
    runtimeParameters:
    - name: TRAINING_RUNS
      value: '100'
    artifacts:
    - name: serving-model
      path: 'Pusher:pushed_model:0[pushed == 1]'
  triggers:
    schedules:
    - cronExpression: '0 * * * *'
      startTime: "2024-01-01T00:00:00Z"
      endTime: "2024-12-31T23:59:59Z"
    onChange:
    - pipeline
    runConfigurations:
    - dependency-rc
```

A Run Configuration can have one of more triggers that determine when the next training run will be started.

## Fields

| Name                                | Description                                                                                                                                                                   |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `spec.run`                          | Definition of any runs created under this run configuration. See [Runs](../run/#fields) for more details.                                                                     |
| `spec.triggers.schedules[]`         | List of schedules for when the runs should be created. See [Schedule Definition](#schedule-definition) for more information.                                                  |
| `spec.triggers.onChange[]`          | Resource attributes that execute training runs. `pipeline` triggers when the referenced pipeline changes. `runSpec` triggers when this resource's spec.run field has changed. |
| `spec.triggers.runConfigurations[]` | RunConfigurations to watch for completion - a run for this RunConfiguration will start every time any of the listed dependencies has finished a run successfully.             |

### Schedule Definition
| Name             | Description                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cronExpression` | Cron expression to execute training runs. It can have 5 (standard cron) or 6 (first digit expresses seconds) fields. When a provider does not support the 6-field format, seconds will be omitted.                                                                                                                                                                                                                            |
| `startTime`      | Optional. If supported by the provider, this is a timestamp after which the first run can be scheduled. Defaults to Schedule create time if not specified.                                                                                                                                                                                                                                                                    |
| `endTime`        | Optional. If supported by the provider, this is a timestamp after which no new runs can be scheduled. If specified, The schedule will be completed when `endTime` is reached. If not specified, new runs will keep getting scheduled until this Schedule is paused or deleted. Already scheduled runs will be allowed to complete. `endTime` must be after `startTime` and the current time in order for this to take effect. |
