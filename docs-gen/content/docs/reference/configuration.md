---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name      | Description                                                                                                                                                                                                                        | Example                                                |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `defaultExperiment` | Default Experiment name to be used for creating pipeline runs                                                                                                                                                                      | `Default`                                              |
| `defaultProvider`   | Default provider name to be used (see [Using Multiple Providers](../providers))                                                                                                                                                    | `vertex-ai-europe`                                     |
| `multiversion`      | If enabled, it will support previous versions of the CRDs, only the latest otherwise                                                                                                                                               | `true`                                                 |
| `workflowNamespace` | Namespace where operator Argo workflows should be running - defaults to the operator's namespace                                                                                                                                   | `kfp-operator-workflows`                               |
| `runCompletionTTL`  | Duration string for how long to keep one-off runs after completion - a zero-length or negative duration will result in runs being deleted immediately after completion; defaults to empty (never delete runs)                      | `10m`                                                  |

An example can be found [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/controller_manager_config.yaml).

## Provider Configurations

The provider configurations are specific to the implementation. The operator supports the following out of the box.

### Common

| Parameter name               | Description                                                                                                                                                                                                                        | Example                                                |
|------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `image`<sup>*</sup>          | Container image of the provider                                                                                                                                                                                                    | `kfp-operator-kfp-provider:0.0.2`                      |
| `executionMode`<sup>*</sup>  | KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode)                                                                                               | `v1` (currently KFP) or `v2` (Vertex AI)               |
| `serviceAccount`<sup>*</sup> | Service Account name to be used for all provider-specific operations (see respective provider)                                                                                                                                     | `kfp-operator-vertex-ai`                               |
| `defaultBeamArgs`            | Default Beam arguments to which the pipeline-defined ones will be added                                                                                                                                                            | <pre>- name: project<br/>  value: my-gcp-project</pre> |
| `pipelineRootStorage`        | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline | `gcs://kubeflow-pipelines-bucket`                      |

<sup>*</sup> field automatically populated by Helm based on provider type

### Kubeflow Pipelines

| Parameter name             | Description                                      | Example                                         |
|----------------------------|--------------------------------------------------|-------------------------------------------------|
| `kfpNamespace`             | The KFP namespace                                | `kubeflow`                                      |
| `restKfpApiUrl`            | The KFP REST URL available to the operator       | `http://ml-pipeline.kubeflow:8888`              |
| `grpcKfpApiAddress`        | The KFP gRPC address for the eventsource server  | `ml-pipeline.kubeflow-pipelines:8887`           |
| `grpcMetadataStoreAddress` | The MLMD gRPC address for the eventsource server | `metadata-grpc-service.kubeflow-pipelines:8080` |

KFP must be installed in [standalone mode](https://www.kubeflow.org/docs/components/pipelines/installation/standalone-deployment/). Default endpoints are used below.

### Vertex AI Pipelines

![Vertex AI Provider](/images/vai-provider.png)

| Parameter name                          | Description                                                                                                          | Example                                                           |
|-----------------------------------------|----------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| `pipelineBucket`                        | GCS bucket where to store the compiled pipeline                                                                      | `kfp-operator-pipelines`                                          |
| `vaiProject`                            | Vertex AI GCP project name                                                                                           | `kfp-operator-vertex-ai`                                          |
| `vaiLocation`                           | Vertex AI GCP project location                                                                                       | `europe-west2`                                                    |
| `vaiJobServiceAccount`                  | Vertex AI GCP service account to run pipeline jobs                                                                   | `kfp-operator-vai@kfp-operator-vertex-ai.iam.gserviceaccount.com` |
| `runsTopic`                             | Pub/Sub topic name to publish runs                                                                                   | `kfp-operator-runs`                                               |
| `submitterRunsSubscription`             | Subscription on the runs topic for the pipeline job submitter                                                        | `kfp-operator-runs-submitter`                                     |
| `eventsourcePipelineEventsSubscription` | Subscription for the eventsource to use which subscribes to the Vertex AI pipeline events log sink topic (see below) | `kfp-operator-vai-run-events-eventsource`                         |
| `maxConcurrentRunCount`                 | Maximum number of runs that can be started concurrently per schedule, defaults to 10                                 | `3`                                                               |

#### GCP Project Setup

The following GCP APIs need to be enabled in the configured `vaiProject`:
- Vertex AI
- Pub/Sub
- Cloud Storage
- Cloud Scheduler

A [Vertex AI log](https://cloud.google.com/vertex-ai/docs/pipelines/logging) sink needs to be created that:
- captures pipeline state changes as
  ```resource.type="aiplatform.googleapis.com/PipelineJob"
     jsonPayload.state="PIPELINE_STATE_SUCCEEDED" OR "PIPELINE_STATE_FAILED" OR "PIPELINE_STATE_CANCELLED"```
- writes state changes to Pub/Sub on to a Pipeline Events topic (see below for required subscription)

Pub/Sub topics and subscriptions need to be created for:
- Runs
  - Topic: `runsTopic`
  - Subscriptions:`submitterRunsSubscription`, `eventsourceRunsSubscription`<sup>*</sup>
- Pipeline Events
  - Subscription: `eventsourcePipelineEventsSubscription`

It is important to configure the retry policy for the `eventsourcePipelineEventsSubscription` subscription according to your needs. This determines the retry frequency of the eventsource server to query the Vertex AI API in case of errors.
We suggest an exponential backoff with min and max backoff set to at least 10 seconds each, resulting in a fixed 10 seconds wait between polls.

GCS pipeline storage bucket `provider.configuration.pipelineBucket` needs to be created

The configured `serviceAccount` needs to have [workload identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and be granted the following permissions:
  - `storage.objects.create` on the configured `pipelineBucket`
  - `storage.objects.get` on the configured `pipelineBucket`
  - `storage.objects.delete` on the configured `pipelineBucket`
  - deprecated - `cloudscheduler.jobs.create`
  - deprecated - `cloudscheduler.jobs.update`
  - deprecated - `cloudscheduler.jobs.delete`
  - deprecated - `projects.topics.publish` to the configured `runs` and `runIntentsTopic` topic
  - `projects.subscriptions.pull` from the configured `enqueuerRunIntentsSubscription`, `submitterRunsSubscription` and `eventsourceRunsSubscription`<sup>*</sup> subscriptions
  - `aiplatform.pipelineJobs.create`
  - `aiplatform.pipelineJobs.get`<sup>*</sup>
  - `aiplatform.schedules.get`
  - `aiplatform.schedules.create`
  - `aiplatform.schedules.delete`
  - `aiplatform.schedules.update`
  - `iam.serviceAccounts.actAs` the configured `vaiJobServiceAccount` Vertex AI Job Runner

<sup>*</sup> fields only needed if the operator is installed with [eventing support](../../getting-started/overview/#eventing-support)
