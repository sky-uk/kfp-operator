---
title: "Provider"
weight: 5
---

The Provider resource represents the provider specific configuration required to submit / update / delete ml resources with the given provider.
e.g Kubeflow Pipelines or the Vertex AI Platform.
Providers configuration can be set using this resource and permissions for access can be configured via service accounts.

> Note: changing the provider of a resource that was previously managed by another provider will result in a resource error.
Any referenced resources must always match the provider of the referencing resource (e.g. RunConfiguration to Pipeline) as updates are not propagated or checked and will result in runtime errors on the provider.

### Common Fields

| Name                       | Description                                                                                                                                                                                                                        | Example                                                |
|----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `spec.serviceImage`        | Container image of [the provider service](../../providers/#provider-service)                                                                                                                                                       | `kfp-operator-kfp-provider-service:0.0.2`              |
| `spec.image`               | Container image of [the provider CLI](../../providers/#provider-cli)                                                                                                                                                               | `kfp-operator-kfp-provider:0.0.2`                      |
| `spec.executionMode`       | KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode)                                                                                               | `v1` (currently KFP) or `v2` (Vertex AI)               |
| `spec.serviceAccount`      | Service Account name to be used for all provider-specific operations (see respective provider)                                                                                                                                     | `kfp-operator-vertex-ai`                               |
| `spec.defaultBeamArgs`     | Default Beam arguments to which the pipeline-defined ones will be added                                                                                                                                                            | <pre>- name: project<br/>  value: my-gcp-project</pre> |
| `spec.pipelineRootStorage` | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline | `gcs://kubeflow-pipelines-bucket`                      |
| `spec.parameters`          | Parameters specific to each provider, i.e. [KFP](#kubeflow-specific-parameters) and [VAI](#vertex-ai-specific-parameters)                                                                                                          | `gcs://kubeflow-pipelines-bucket`                      |

### Kubeflow:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: kfp
  namespace: kfp-operator
spec:
  serviceImage: kfp-operator-kfp-provider-service:<version>
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
| `parameters.grpcKfpApiAddress`        | The exposed grpc endpoint used to interact with Kubeflow pipelines        |
| `parameters.grpcMetadataStoreAddress` | The exposed grpc endpoint used for metadata store with Kubeflow pipelines |
| `parameters.kfpNamespace`             | The namespace where Kubeflow is deployed                                  |
| `parameters.restKfpApiUrl`            | The exposed restful endpoint used to interact with Kubeflow pipelines     |


### Vertex AI:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: vai
  namespace: kfp-operator
spec:
  serviceImage: kfp-operator-vai-provider-service:<version>
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
| `parameters.eventsourcePipelineEventsSubscription` | The eventsource subscription used to capture run-completion events   |
| `parameters.maxConcurrentRunCount`                 | The number of pipelines that may run concurrently                    |
| `parameters.pipelineBucket`                        | The output storage bucket for a trained pipeline model               |
| `parameters.vaiJobServiceAccount`                  | The service account should be used by VAI when submitting a pipeline |
| `parameters.vaiLocation`                           | The region VAI should run a pipeline within                          |
| `parameters.vaiProject`                            | The project VAI should run a pipeline within                         |
