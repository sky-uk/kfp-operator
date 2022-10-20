---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name       | Description                                                                                                                                                                                                                        | Example                                                |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `defaultBeamArgs`    | Default Beam arguments to which the pipeline-defined ones will be added                                                                                                                                                            | <pre>- name: project<br/>  value: my-gcp-project</pre> |
| `defaultExperiment`  | Default Experiment name to be used for creating pipeline runs                                                                                                                                                                      | `Default`                                              |
| `debug`              | Default debugging options                                                                                                                                                                                                          | See [Debugging](../debugging)                          |
| `multiversion`       | If enabled, it will support previous versions of the CRDs, only the latest otherwise                                                                                                                                               | `true`                                                 |
| `pipelineStorage`    | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline | `gcs://kubeflow-pipelines-bucket`                      |
| `providerConfigFile` | File in the controller container containing the provider-specific configuration (see below)                                                                                                                                        | `provider.yaml`                                        |

An example can be found in the [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/controller_manager_config.yaml).

## Provider Configuration

The provider configuration is specific to the implementation. The operator supports the following out of the box.

An example can be found [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/provider.yaml).

### Common

| Parameter name  | Description                                                                                                                          | Example                                  |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------|
| `image`         | Container image of the provider                                                                                                      | `kfp-operator-kfp-provider:0.0.2`        |
| `executionMode` | KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode) | `v1` (currently KFP) or `v2` (Vertex AI) |

### Kubeflow Pipelines

| Parameter name | Description                                | Example                               |
|----------------|--------------------------------------------|---------------------------------------|
| `endpoint`     | The KFP endpoint available to the operator | `kubeflow-ui.kubeflow-pipelines:8080` |

### Vertex AI Pipelines

| Parameter name   | Description                                     | Example                  |
|------------------|-------------------------------------------------|--------------------------|
| `pipelineBucket` | GCS bucket where to store the compiled pipeline | `kfp-operator-pipelines` |
