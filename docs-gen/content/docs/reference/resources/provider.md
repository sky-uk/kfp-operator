---
title: "Provider"
weight: 6
---

The Provider resource represents the provider specific configuration required to submit / update / delete ml resources with the given provider.
e.g Vertex AI Platform.
Providers configuration can be set using this resource and permissions for access can be configured via service accounts.

> Note: changing the provider of a resource that was previously managed by another provider will result in a resource error.
Any referenced resources must always match the provider of the referencing resource (e.g. RunConfiguration to Pipeline) as updates are not propagated or checked and will result in runtime errors on the provider.

### Common Fields

| Name                                    | Description                                                                                                                                                                                                                                                                                                   | Example                                                                                                             |
| :-------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | :------------------------------------------------------------------------------------------------------------------ |
| `spec.serviceImage`                     | Container image of [the provider service](../../providers/#provider-service)                                                                                                                                                                                                                                  | `kfp-operator-vai-provider-service:0.0.2`                                                                           |
| `spec.executionMode`                    | Deprecated: This field will be dropped in future versions and the provider will only support v2 execution mode under the hood. KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode)                                           |                                                                                                                     |
| `spec.serviceAccount`                   | Service Account name to be used for all provider-specific operations (see respective provider)                                                                                                                                                                                                                | `kfp-operator-vertex-ai`                                                                                            |
| `spec.pipelineRootStorage`              | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline                                                                            | `gcs://bucket`                                                                                   |
| `spec.parameters`                       | Parameters specific to each provider, e.g. [VAI](#vertex-ai-specific-parameters)                                                                                                                                                                                     | `gcs://bucket`                                                                                   |
| `spec.frameworks`                       | Frameworks supported by the provider. Currently only `tfx` is supported.                                                                                                                                                                                                                                      |                                                                                                                     |
| `spec.frameworks[0].name`               | Name of the framework.                                                                                                                                                                                                                                                                                        | `tfx`                                                                                                               |
| `spec.frameworks[0].image`              | Framework image.                                                                                                                                                                                                                                                                                              | `ghcr.io/kfp-operator/kfp-operator-tfx-compiler:version-tag`                                                        |
| `spec.frameworks[0].patches`            | List of JSON patches that will be applied to every `Pipeline` resource that uses this `Provider` before it's passed (as JSON) to the corresponding Argo Workflow for processing.                                                                                                                              |                                                                                                                     |
| `spec.frameworks[0].patches[0].type`    | The type of patch to be applied to the pipeline resource definition JSON. Can be either `json` ([RFC6902](https://datatracker.ietf.org/doc/html/rfc6902)) or `merge` ([RFC7396](https://datatracker.ietf.org/doc/html/rfc7396)).                                                                              | `json`                                                                                                              |
| `spec.frameworks[0].patches[0].payload` | The patch to be applied to the pipeline resource definition JSON.                                                                                                                                                                                                                                             | `[{ "op": "add", "path": "/framework/parameters/beamArgs/0", "value": { "name": "newArg", "value": "newValue" } }]` |
| `spec.allowedNamespaces`                | A list of namespaces that resources can reference this provider from. If a resource tries to reference this provider from a namespace not in the `allowedNamespaces` list, the resource will fail. If no allowedNamespaces list is configured, then resources can reference this provider from any namespace. | ```- default ```                                                                                                    |


### Vertex AI:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: vai
  namespace: kfp-operator
spec:
  serviceImage: kfp-operator-vai-provider-service:<version>
  pipelineRootStorage: gs://<storage_location>
  serviceAccount: kfp-operator-vai
  parameters:
    eventsourcePipelineEventsSubscription: kfp-operator-vai-run-events-eventsource
    maxConcurrentRunCount: 1
    pipelineBucket: pipeline-storage-bucket
    vaiJobServiceAccount: kfp-operator-vai@<project>.iam.gserviceaccount.com
    vaiLocation: europe-west2
    vaiProject: <project>
  frameworks:
  - name: tfx
    image: ghcr.io/kfp-operator/kfp-operator-tfx-compiler:version-tag
    patches:
    - type: json
      patch: |
        [
          {
            "op": "add",
            "path": "/framework/parameters/beamArgs/0",
            "value": {
              "name": "project",
              "value": "<project>"
            }
          }
        ]
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
