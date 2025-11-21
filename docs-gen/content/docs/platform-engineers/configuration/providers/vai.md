---
title: "Vertex AI"
---

## Overview

The **Vertex AI (VAI) Provider Service** is included within this project to interface directly with 
Google Cloud's Vertex AI platform. This service acts as a bridge between the KFP Operator and Vertex AI, enabling 
seamless management and execution of machine learning workflows.

![VAI Provider]({{< param "subpath" >}}/master/images/vai.svg)

> Note: VAI does not support the `experiment` resource

## Deployment and Usage

Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Vertex AI instance. [See the getting started guide.](../../../getting-started/installation/#providers)

The configuration can be managed via the [provider custom resource](../../resources/provider/#vertex-ai) installed by the operator.

In order for eventing to be configured for VAI, some work is required to export logs from Vertex AI to pubsub for the 
provider service to consume. [Instructions on how to do this can be found here.](#gcp-project-setup)

For detailed implementation code and further technical insights, refer to the
[VAI Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/vai) in the
repository.

### GCP Project Setup

The following GCP APIs need to be enabled in the configured `vaiProject`:
- Vertex AI
- Pub/Sub
- Cloud Storage
- Cloud Scheduler

#### Log Sink
A [Vertex AI log](https://cloud.google.com/vertex-ai/docs/pipelines/logging) sink needs to be created that:
- captures pipeline state changes as
  ```resource.type="aiplatform.googleapis.com/PipelineJob" jsonPayload.state="PIPELINE_STATE_SUCCEEDED" OR "PIPELINE_STATE_FAILED" OR "PIPELINE_STATE_CANCELLED"```
- writes state changes to Pub/Sub on to a Pipeline Events topic (see below for required subscription)

#### Pub Sub
Pub/Sub topics and subscriptions need to be created for:
- Pipeline Events
    - Subscription: `eventsourcePipelineEventsSubscription`

It is important to configure the retry policy for the `eventsourcePipelineEventsSubscription` subscription according to your needs. This determines the retry frequency of the eventsource server to query the Vertex AI API in case of errors.
We suggest an exponential backoff with min and max backoff set to at least 10 seconds each, resulting in a fixed 10 seconds wait between polls.

GCS pipeline storage bucket `provider.configuration.pipelineBucket` needs to be created

#### RBAC
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

<sup>*</sup> fields only needed if the operator is installed with [eventing support](../../../getting-started/overview/#eventing-support)


## Implementation Details

- **API**: Implements the [openapi spec for provider services](../overview/#api). 
- **Event Handling**: The events are sourced from Pubsub where log output from Vertex AI is formatted in such a way to be consumed by the provider service as 
a `run completion event`. This is then processed and sent to the operators webhook to update the status of the run.
