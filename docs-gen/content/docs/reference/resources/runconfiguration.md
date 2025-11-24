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
    provider: provider-namespace/provider-name
    pipeline: penguin-pipeline
    experimentName: penguin-experiment
    parameters:
    - name: TRAINING_RUNS
      value: '100'
    - name: push_destination
      value: '{"filesystem":{"base_directory":"gs://my-bucket/penguin-pipeline"}}'
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
    - base-namespace/dependency-rc
```

A Run Configuration can have one of more triggers that determine when the next training run will be started.

## Fields

| Name                                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                         |
|-------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.run`                          | Definition of any runs created under this run configuration. See [Runs](../run/#fields) for more details.                                                                                                                                                                                                                                                                                                                                           |
| `spec.triggers.schedules[]`         | List of schedules for when the runs should be created. See [Schedule Definition](../runschedule/#schedule-definition) for more information.                                                                                                                                                                                                                                                                                                         |
| `spec.triggers.onChange[]`          | Resource attributes that execute training runs. `pipeline` triggers when the referenced pipeline changes. `runSpec` triggers when this resource's spec.run field has changed.                                                                                                                                                                                                                                                                       |
| `spec.triggers.runConfigurations[]` | RunConfigurations to watch for completion - a run for this RunConfiguration will start every time any of the listed dependencies has finished a run successfully. RunConfigurations in other namespaces can trigger this RunConfiguration by using the format `namespace/runConfigurationName`. If no namespace is set, the operator will assume the RunConfiguration being watched is in the same namespace as the RunConfiguration being applied. |
