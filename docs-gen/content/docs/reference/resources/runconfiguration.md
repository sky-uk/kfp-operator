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

| Name                         | Description                                                                                                                                                                                                                                       |
|------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.run.pipeline`          | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will track the latest version of the specified pipeline.                                           |
| `spec.run.experimentName`    | The name of the corresponding experiment resource (optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided). |
| `spec.run.artifacts[]`       | Exposed models that will be included in run completion event when a run for this RunConfiguration has succeeded. See [Run Artifact Definition](../run/#run-artifact-definition) for more information.                                             | 
| `spec.run.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline.                                                                                                                                                                                 |
| `spec.triggers.schedules[]`  | Cron schedules to execute training runs. It can have 5 (standard cron) or 6 (first digit expresses seconds) fields. When a provider does not support the 6-field format, seconds will be omitted.                                                 |
| `spec.triggers.onChange[]`   | Resource attributes that execute training runs. Currently, only `pipeline` is supported, triggering when the referenced pipeline changes.                                                                                                         |
