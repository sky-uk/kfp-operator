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
    artifacts:
    - name: serving-model
      path: 'Pusher:pushed_model:0[pushed == 1]'
  triggers:
    schedules:
    - '0 * * * *'
    onChange:
    - pipeline
```

A Run Configuration can have one of more triggers that determine when the next training run will be started.

## Fields

| Name                        | Description                                                                                                                                                                                       |
|-----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.run`                  | Definition of any runs created under this run configuration. See [Runs](../run/#fields) for more details.                                                                                         |
| `spec.triggers.schedules[]` | Cron schedules to execute training runs. It can have 5 (standard cron) or 6 (first digit expresses seconds) fields. When a provider does not support the 6-field format, seconds will be omitted. |
| `spec.triggers.onChange[]`  | Resource attributes that execute training runs. Currently, only `pipeline` is supported, triggering when the referenced pipeline changes.                                                         |
