---
title: "Run"
weight: 3
---

The Run resource represents the lifecycle of a one-off run.
One-off pipeline training runs can be configured using this resource as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Run
metadata:
  generateName: penguin-pipeline-run-
spec:
  provider: provider-namespace/provider-name
  pipeline: penguin-pipeline
  experimentName: penguin-experiment
  parameters:
  - name: TRAINING_RUNS
    value: '100'
  - name: EXAMPLES
    valueFrom:
      runConfigurationRef:
        name: base-namespace/penguin-pipeline-example-generator-runconfiguration
        outputArtifact: examples
  artifacts:
  - name: serving-model
    path: 'Pusher:pushed_model:0[pushed == 1]'
```

Note the usage of `metadata.generateName` which tells Kubernetes to generate a new name based on the given prefix for every new resource.
> In general, we expect users to deploy [RunConfigurations](../runconfiguration) to configure the lifecycle of their runs, leaving the management of `Runs` to the operator.

## Fields

| Name                   | Description                                                                                                                                                                                                                                       |
|------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `spec.provider`        | The namespace and name of the associated [Provider resource](../provider/) separated by a `/`, e.g. `provider-namespace/provider-name`.                                                                                                           |
| `spec.pipeline`        | The [identifier](../pipeline/#identifier) of the corresponding pipeline resource to run. If no version is specified, then the RunConfiguration will use the latest version of the specified pipeline.                                             |
| `spec.experimentName`  | The name of the corresponding experiment resource (optional - the `Default` Experiment as defined in the [Installation and Configuration section of the documentation](README.md#configuration) will be used if no `experimentName` is provided). |
| `spec.parameters[]`    | Parameters for the pipeline training run. [See Run Parameters](#run-parameters-definition).                                                                                                                                                       |
| `spec.run.artifacts[]` | Exposed output artifacts that will be included in run completion event when this run has succeeded. See below for more information.                                                                                                               |

### Run Parameters Definition

A pipeline run can be parameterised using parameters.

| Name                                           | Description                                                                                                                                                                                                                                                               |
|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`                                         | The name of the runtime parameter as referenced by the pipeline.                                                                                                                                                                                                          |
| `value`                                        | The value of the runtime parameter.                                                                                                                                                                                                                                       |
| `valueFrom.runConfigurationRef`                | If set, the value of this runtime parameter will be resolved from the output artifacts of the referenced runconfiguration and updated on change.                                                                                                                          |
| `valueFrom.runConfigurationRef.name`           | The namespace and name of the RunConfiguration to resolve in the format `namespace/runConfigurationName`. If no namespace is set, the operator assumes the RunConfiguration to resolve is in the same namespace as the RunConfiguration being applied.                    |
| `valueFrom.runConfigurationRef.outputArtifact` | The name of the outputArtifact to resolve.                                                                                                                                                                                                                                |
| `valueFrom.runConfigurationRef.optional`       | Whether or not the resolution of this parameter is optional. If set to true, and the `outputArtifact` being referenced cannot be found, a run will be created without this parameter. If set to false or not set, no run will be created as the artifact cannot be found. |

Note: either `value` or `valueFrom` must be defined.

### Run Artifact Definition

A pipeline run can expose what Artifacts to include in resulting run completion events. 

| Name   | Description                                                                           |
|--------|---------------------------------------------------------------------------------------|
| `name` | The name to be used in run completion events or references to identify this artifact. |
| `path` | Path of the artifact in the pipeline graph. See below for the syntax                  |

Artifact path Syntax: `<COMPONENT>:<OUTPUT>:<INDEX>[<FILTER>]` with the following parts:

| Part      | Description                                                                        | Example      |
|-----------|------------------------------------------------------------------------------------|--------------|
| COMPONENT | The Pipeline component that produces the artifacts                                 | Pusher       |
| OUTPUT    | The output artifact name of the component                                          | pushed_model |
| INDEX     | The artifact index, defaults to 0 as in most cases there will be only one artifact | 0            |
| FILTER    | A boolean expression to apply to properties of the artifact, defaults to no filter | pushed == 1  |

## Lifecycle

The KFP-Operator tracks the completion of the created run in the `CompletionState` of the resource's status.
The operator will clean up completed runs automatically based on the configured TTL. See [Configuration](../../platform-engineers/configuration/operator-configuration) for more information.
