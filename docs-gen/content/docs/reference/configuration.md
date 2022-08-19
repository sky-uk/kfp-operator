---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name      | Description                                                                                                                         | Example                               |
|---------------------|-------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------|
| `defaultBeamArgs`   | Default Beam arguments on which the pipeline-defined ones will be overlaid                                                          | `project: my-gcp-project`             |
| `defaultExperiment` | Default Experiment name to be used for creating pipeline runs                                                                       | `Default`                             |
| `debug`             | Default debugging options                                                                                                           | See [Debugging](./debugging.md)       |
| `kfpEndpoint`       | The KFP endpoint available to the operator                                                                                          | `kubeflow-ui.kubeflow-pipelines:8080` |
| `pipelineStorage`   | The storage location used by [TFX](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs | `gcs://kubeflow-pipelines-bucket`     |

An example can be found in the [local run configuration](../config/manager/controller_manager_config.yaml).
