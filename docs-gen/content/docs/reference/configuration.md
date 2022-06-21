---
title: "Configuration"
---

### Configuration

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name | Description | Example |
| --- | --- | --- |
| `argo.compilerImage` | The KFP Operator compiler image | `docker.io/kfp-operator-argo-compiler:abcdef` |
| `argo.containerDefaults` | Container Spec defaults to be used for Argo workflow pods created by the operator | `{}` |
| `argo.kfpSdkImage` | The KFP Operator tools image | `docker.io/kfp-operator-argo-kfp-sdk:abcdef` |
| `argo.metadataDefaults` | Container Metadata defaults to be used for Argo workflow pods created by the operator | `{}` |
| `argo.serviceAccount` | The [k8s Service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) used to run argo workflows | `kfp-operator-sa` |
| `defaultBeamArgs` | Default Beam arguments on which the pipeline-defined ones will be overlaid | `project: my-gcp-project` |
| `defaultExperiment` | Default Experiment name to be used for creating pipeline runs | `Default` |
| `debug` | Default debugging options | See [Debugging](./debugging.md) |
| `kfpEndpoint` | The KFP endpoint available to the operator | `kubeflow-ui.kubeflow-pipelines:8080` |
| `pipelineStorage` | The storage location used by [TFX](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs | `gcs://kubeflow-pipelines-bucket` |

An example can be found in the [local run configuration](../config/manager/controller_manager_config.yaml).
