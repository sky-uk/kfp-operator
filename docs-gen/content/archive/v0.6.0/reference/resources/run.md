---
title: "Run"
weight: 3
---

The Run resource represents the lifecycle of a one-off run.
One-off pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Run
metadata:
  generateName: penguin-pipeline-run-
spec:
  pipeline: penguin-pipeline:v1-abcdef
  experimentName: penguin-experiment
  runtimeParameters:
  - name: TRAINING_RUNS
    value: '100'
  - name: EXAMPLES
    valueFrom:
      runConfigurationRef:
        name: penguin-pipeline-example-generator-runconfiguration
        outputArtifact: examples
  artifacts:
  - name: serving-model
    path: 'Pusher:pushed_model:0[pushed == 1]'
```

Note the usage of `metadata.generateName` which tells Kubernetes to generate a new name based on the given prefix for every new resource.

## Fields

| Name                       | Description                                                                                                                                                                                                                                       |
|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.pipeline`            | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will use the latest version of the specified pipeline.                                             |
| `spec.experimentName`      | The name of the corresponding experiment resource (optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided). |
| `spec.runtimeParameters[]` | Runtime parameters for the pipeline training run. See below for more information.                                                                                                                                                                 |
| `spec.run.artifacts[]`     | Exposed output artifacts that will be included in run completion event when this run has succeeded. See below for more information.                                                                                                               |

### Runtime Parameter Definition

A pipeline run can be parameterised using RunTimeParameters.

| Name                                           | Description                                                                                                                                      |
|------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`                                         | The name of the runtime parameter as referenced by the pipeline.                                                                                 |
| `value`                                        | The value of the runtime parameter.                                                                                                              |
| `valueFrom.runConfigurationRef`                | If set, the value of this runtime parameter will be resolved from the output artifacts of the referenced runconfiguration and updated on change. |
| `valueFrom.runConfigurationRef.name`           | The name of the runconfiguration to resolve.                                                                                                     |
| `valueFrom.runConfigurationRef.outputArtifact` | The name of the outputArtifact to resolve.                                                                                                       |

Note: either `value` or `valueFrom` must be defined.

### Run Artifact Definition

A pipeline run can expose what Artifacts to include in resulting run completion events. 

| Name     | Description                                                                           |
|----------|---------------------------------------------------------------------------------------|
| `name`   | The name to be used in run completion events or references to identify this artifact. |
| `path`   | Path of the artifact in the pipeline graph. See below for the syntax                  |

Artifact path Syntax: `<COMPONENT>:<OUTPUT>:<INDEX>[<FILTER>]` with the following parts:

| Part      | Description                                                                        | Example      |
|-----------|------------------------------------------------------------------------------------|--------------|
| COMPONENT | The Pipeline component that produces the artifacts                                 | Pusher       |
| OUTPUT    | The output artifact name of the component                                          | pushed_model |
| INDEX     | The artifact index, defaults to 0 as in most cases there will be only one artifact | 0            |
| FILTER    | A boolean expression to apply to properties of the artifact, defaults to no filter | pushed == 1  |

## Lifecycle

The KFP-Operator tracks the completion of the created run in the `CompletionState` of the resource's status.
The operator will clean up completed runs automatically based on the configured TTL. See [Configuration](../../configuration) for more information.
