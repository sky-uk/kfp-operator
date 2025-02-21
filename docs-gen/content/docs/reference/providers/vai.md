---
title: "Vertex AI"
---

## Overview

The KFP Operator's **Vertex AI (VAI) Provider Service** is included within this project to interface directly with 
Google Cloud's Vertex AI platform. This service acts as a bridge between the KFP Operator and Vertex AI, enabling 
seamless management and execution of machine learning workflows.

![VAI Provider]({{< param "subpath" >}}/master/images/vai.svg)

> Note: VAI does not support the `experiment` resource

## Deployment and Usage

Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Vertex AI instance. [See the getting started guide.](../../../getting-started/installation/#providers)

The configuration can be managed via the [provider custom resource](../../resources/provider/#vertex-ai) installed by the operator.

In order for eventing to be configured for VAI, some work is required to export logs from Vertex AI to pubsub for the 
provider service to consume. [Instructions on how to do this can be found here.](../../configuration/#gcp-project-setup)

For detailed implementation code and further technical insights, refer to the
[VAI Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/vai) in the
repository.

## Implementation Details

- **API**: Implements the [openapi spec for provider services](../overview/#api). 
- **Event Handling**: The events are sourced from Pubsub where log output from Vertex AI is formatted in such a way to be consumed by the provider service as 
a `run completion event`. This is then processed and sent to the operators webhook to update the status of the run.
