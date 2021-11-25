# Documentation

## Installation

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

### Helm Chart

The operator can be installed using helm by providing a valid `values.yaml` file to override the [defaults](../helm/kfp-operator/values.yaml). 

```
make helm-install
```

When installing with Helm, the above configuration can be set in the `configuration` section of `values.yaml`.
Note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied.

## TFX Pipelines and Components

Unlike imparative Kubeflow Pipelines deployments, the operator takes care of providing all environment-specific configuration and setup for the pipelines. Pipeline creators therefore don't have to provide DAG runners, metadata configs, serving directories, etc. Furthermore, pusher is not required and the operator can extend the pipeline with this very environment-specific component.

For running a pipeline using the operator, only the list of TFX components needs to be returned. Everything else is done by the operator. See the [penguin pipeline](./quickstart/penguin_pipeline/pipeline.py) for an example.

### Lifecycle phases and Parameter types

TFX Pipelines go through certain lifecycle phases that are unique to this technology. It is helpful to understand where these differ and where they are executed.

**Development:** Creating the components definition as code.

**Compilation:** Applying compile-time parameters and defining the execution runtime (aka DAG runner) for the pipeline to be compiled into a deployable artifact.

**Deployment:** Creating a pipeline representation in the target environment.

**Running:** Instantiating the pipeline, applying runtime parameters and running all pipeline steps involved to completion.

*Note:* Local runners usually skip compilation and deployment and run the pipeline straight away.

TFX allows the parameterisation of Pipelines in most lifecycle stages:

| Parameter type | Description | Example |
| --- | --- | --- |
| Named Constants | Code constants | ANN layer size |
| Compile-time parameter | Parameters that are unlikely to change between pipeline runs supplied as environment variabels to the pipeline function | Bigquery dataset |
| Runtime parameter | Parameters exposed as TFX [RuntimeParameter](https://www.tensorflow.org/tfx/api_docs/python/tfx/v1/dsl/experimental/RuntimeParameter?hl=en) which can be overridden at runtime allow simplified experimentation without having to recompile the pipeline | Number of training runs |

The pipeline operator supports the application of compile time and runtime parameters through its custom resources. We strongly encourage the usage of both of these parameter types to speed up development and experimentation lifecycles. Note that Runtime parameters can initialised to default values from both constants and compile-time parameters

## Operator Resources

The Kubeflow Pipelines operator manages the lifecycle of pipelines and related resources via Kubernetes Resources:

- [Pipeline](pipeline.md)
- [RunConfiguration](runconfiguration.md)