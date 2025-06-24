---
title: "RunSchedule"
weight: 4
---

The RunSchedule resource represents the lifecycle of scheduled runs.

Schedules for pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: RunSchedule
metadata:
  generateName: penguin-pipeline-run-schedule-
spec:
  artifacts:
  - name: serving-model
    path: 'Pusher:pushed_model:0[pushed == 1]'
  experimentName: penguin-experiment
  pipeline: penguin-pipeline

  provider: provider-namespace/provider-name
  parameters:
  - name: TRAINING_RUNS
    value: '100'
  - name: EXAMPLES
    valueFrom:
      runConfigurationRef:
        name: base-namespace/penguin-pipeline-example-generator-runconfiguration
        outputArtifact: examples
  schedule:
    cronExpression: '0 * * * *'
    startTime: "2024-01-01T00:00:00Z"
    endTime: "2024-12-31T23:59:59Z"
```

Note the usage of `metadata.generateName` which tells Kubernetes to generate a new name based on the given prefix for every new resource.
> In general, we expect users to deploy [RunConfigurations](../runconfiguration) to configure the lifecycle of their runs, leaving the management of `RunSchedules` to the operator.

## Fields

| Name                  | Description                                                                                                                                                                                                                                |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `spec.provider`       | The namespace and name of the associated [Provider resource](../provider/) separated by a `/`, e.g. `provider-namespace/provider-name`.                                                                                                    |
| `spec.pipeline`       | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunSchedule will use the latest version of the specified pipeline.                                           |
| `spec.experimentName` | The name of the corresponding experiment resource (optional - the `Default` Experiment as defined in the [Installation section](../../../getting-started/installation#build-and-install) will be used if no `experimentName` is provided). |
| `spec.parameters[]`   | Parameters for the pipeline training run. [See Run Parameters](../run#run-parameters-definition).                                                                                                                                          |
| `spec.artifacts[]`    | Exposed output artifacts that will be included in run completion event when this run has succeeded. See the [Run Artifact Definition](../run#run-artifact-definition) for more detail.                                                     |
| `spec.schedule`       | for when the runs should be created. See [Schedule Definition](#schedule-definition) for more detail.                                                                                                                                      |


### Schedule Definition
| Name             | Description                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cronExpression` | Cron expression to execute training runs. It can have 5 (standard cron) or 6 (first digit expresses seconds) fields. When a provider does not support the 6-field format, seconds will be omitted.                                                                                                                                                                                                                            |
| `startTime`      | Optional. If supported by the provider, this is a timestamp after which the first run can be scheduled. Defaults to Schedule create time if not specified.                                                                                                                                                                                                                                                                    |
| `endTime`        | Optional. If supported by the provider, this is a timestamp after which no new runs can be scheduled. If specified, The schedule will be completed when `endTime` is reached. If not specified, new runs will keep getting scheduled until this Schedule is paused or deleted. Already scheduled runs will be allowed to complete. `endTime` must be after `startTime` and the current time in order for this to take effect. |
