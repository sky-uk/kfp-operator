---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name      | Description                                                                                                                                                                                                   | Example                                        |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|
| `defaultExperiment` | Default Experiment name to be used for creating pipeline runs                                                                                                                                                 | `Default`                                      |
| `defaultProvider`   | Default provider name to be used (see [Using Multiple Providers](../providers))                                                                                                                               | `vertex-ai-europe`                             |
| `multiversion`      | If enabled, it will support previous versions of the CRDs, only the latest otherwise                                                                                                                          | `true`                                         |
| `workflowNamespace` | Namespace where operator Argo workflows should be running - defaults to the operator's namespace                                                                                                              | `kfp-operator-workflows`                       |
| `runCompletionTTL`  | Duration string for how long to keep one-off runs after completion - a zero-length or negative duration will result in runs being deleted immediately after completion; defaults to empty (never delete runs) | `10m`                                          |
| `runCompletionFeed` | Configuration of the service for the run completion feed back to KFP Operator                                                                                                                                 | See [here](#run-completion-feed-configuration) |

An example can be found [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/controller_manager_config.yaml).

## Run Completion Feed Configuration
| Parameter name                | Description                                                        | Example                                                                        |
|-------------------------------|--------------------------------------------------------------------|--------------------------------------------------------------------------------|
| `runCompletionFeed.port`      | The port that the feed endpoint will listen on                     | `8082`                                                                         |
| `runCompletionFeed.endpoints` | Array of upstream endpoints that should be called per feed message | `- host: upstream-service<br/>&nbsp;&nbsp;path: /<br/>&nbsp;&nbsp;port: 12000` |

## Provider Configurations

The provider configurations are specific to the implementation, these configuration are applied via [Provider Custom Resource](../resources/provider). 

### Kubeflow Pipelines

KFP must be installed in [standalone mode](https://www.kubeflow.org/docs/components/pipelines/installation/standalone-deployment/). 
Its configuration can be controlled using the [KFP specific parameters within a Provider Resource](../resources/provider/#kubeflow).

### Vertex AI Pipelines

VAI configuration can be controlled using [VAI specific parameters within a Provider Resource](../resources/provider/#vertex-ai)
![Vertex AI Provider](/images/vai-provider.png)

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
- Pipeline Events
  - Subscription: `eventsourcePipelineEventsSubscription`

It is important to configure the retry policy for the `eventsourcePipelineEventsSubscription` subscription according to your needs. This determines the retry frequency of the eventsource server to query the Vertex AI API in case of errors.
We suggest an exponential backoff with min and max backoff set to at least 10 seconds each, resulting in a fixed 10 seconds wait between polls.

GCS pipeline storage bucket `provider.configuration.pipelineBucket` needs to be created

The configured `serviceAccount` needs to have [workload identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and be granted the following permissions:
  - `storage.objects.create` on the configured `pipelineBucket`
  - `storage.objects.get` on the configured `pipelineBucket`
  - `storage.objects.delete` on the configured `pipelineBucket`
  - `projects.subscriptions.pull` from the configured `eventsourcePipelineEventsSubscription`<sup>*</sup> subscription
  - `aiplatform.pipelineJobs.create`
  - `aiplatform.pipelineJobs.get`<sup>*</sup>
  - `aiplatform.schedules.get`
  - `aiplatform.schedules.create`
  - `aiplatform.schedules.delete`
  - `aiplatform.schedules.update`
  - `iam.serviceAccounts.actAs` the configured `vaiJobServiceAccount` Vertex AI Job Runner

<sup>*</sup> fields only needed if the operator is installed with [eventing support](../../getting-started/overview/#eventing-support)
