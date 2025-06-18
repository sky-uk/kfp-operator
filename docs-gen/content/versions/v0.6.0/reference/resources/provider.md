---
title: "Provider"
weight: 5
---

The Provider resource represents the provider specific configuration required to submit / update / delete ml resources with the given provider.
e.g Kubeflow Pipelines or the Vertex AI Platform.
Providers configuration can be set using this resource and permissions for access can be configured via service accounts.

### Common Fields

| Name                              | Description                                                                                                                                                                                                                        | Example                                                |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| `spec.image`<sup>*</sup>          | Container image of the provider                                                                                                                                                                                                    | `kfp-operator-kfp-provider:0.0.2`                      |
| `spec.executionMode`<sup>*</sup>  | KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode)                                                                                               | `v1` (currently KFP) or `v2` (Vertex AI)               |
| `spec.serviceAccount`<sup>*</sup> | Service Account name to be used for all provider-specific operations (see respective provider)                                                                                                                                     | `kfp-operator-vertex-ai`                               |
| `spec.defaultBeamArgs`            | Default Beam arguments to which the pipeline-defined ones will be added                                                                                                                                                            | <pre>- name: project<br/>  value: my-gcp-project</pre> |
| `spec.pipelineRootStorage`        | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline | `gcs://kubeflow-pipelines-bucket`                      |

<sup>*</sup> field automatically populated by Helm based on provider type

### Kubeflow:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Provider
metadata:
  name: kfp
  namespace: kfp-operator
spec:
  image: kfp-operator-kfp-provider:<version>
  defaultBeamArgs:
  - name: project
    value: <project>
  executionMode: v1
  pipelineRootStorage: gs://<storage_location>
  serviceAccount: kfp-operator-kfp
  parameters:
    grpcKfpApiAddress: ml-pipeline.kubeflow:8887
    grpcMetadataStoreAddress: metadata-grpc-service.kubeflow:8080
    kfpNamespace: kubeflow
    restKfpApiUrl: http://ml-pipeline.kubeflow:8888
```

#### Kubeflow Specific Parameters
| Name                                  | Description                                                               |
| ------------------------------------- | ------------------------------------------------------------------------- |
| `grpcKfpApiAddress`        | The exposed grpc endpoint used to interact with Kubeflow pipelines        |
| `grpcMetadataStoreAddress` | The exposed grpc endpoint used for metadata store with Kubeflow pipelines |
| `kfpNamespace`             | The namespace where Kubeflow is deployed                                  |
| `restKfpApiUrl`            | The exposed restful endpoint used to interact with Kubeflow pipelines     |


### Vertex AI:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Provider
metadata:
  name: vai
  namespace: kfp-operator
spec:
  image: kfp-operator-vai-provider:<version>
  defaultBeamArgs:
  - name: project
    value: <project>
  executionMode: v2
  pipelineRootStorage: gs://<storage_location>
  serviceAccount: kfp-operator-vai
  parameters:
    eventsourcePipelineEventsSubscription: kfp-operator-vai-run-events-eventsource
    maxConcurrentRunCount: 1
    pipelineBucket: pipeline-storage-bucket
    vaiJobServiceAccount: kfp-operator-vai@<project>.iam.gserviceaccount.com
    vaiLocation: europe-west2
    vaiProject: <project>
```

#### Vertex AI Specific Parameters
| Name                                               | Description                                                          |
| -------------------------------------------------- | -------------------------------------------------------------------- |
| `eventsourcePipelineEventsSubscription` | The eventsource subscription used to capture run-completion events   |
| `maxConcurrentRunCount`                 | The number of pipelines that may run concurrently                    |
| `pipelineBucket`                        | The output storage bucket for a trained pipeline model               |
| `vaiJobServiceAccount`                  | The service account should be used by VAI when submitting a pipeline |
| `vaiLocation`                           | The region VAI should run a pipeline within                          |
| `vaiProject`                            | The project VAI should run a pipeline within                         |
