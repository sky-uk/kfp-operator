---
title: "Run"
weight: 3
---

The Run resource represents the lifecycle of a one-off run.
One-off pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha4
kind: Run
metadata:
    generateName: penguin-pipeline-run-
spec:
    pipeline: penguin-pipeline:v1-abcdef
    experimentName: penguin-experiment
    runtimeParameters:
    - name: TRAINING_RUNS
      value: '100'
```

Note the usage of `metadata.generateName` which tells Kubernetes to generate a new name based on the given prefix for every new resource.

## Fields

| Name                     | Description                                                                                                                                                                                                                                      |
|--------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.pipeline`          | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will track the latest version of the specified pipeline.                                          |
| `spec.experimentName`    | The name of the corresponding experiment resource (Optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided) |
| `spec.runtimeParameters` | Dictionary of runtime-time parameters as exposed by the pipeline                                                                                                                                                                                 |

## Lifecycle

The KFP-Operator tracks the completion of the created run in the `CompletionState` of the resource's status.
The operator will clean up completed runs automatically based on the configured TTL. See [Configuration](../../configuration) for more information.