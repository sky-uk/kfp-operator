---
title: "Kubeflow Pipelines"
---

## Overview

The KFP Operator's **Kubeflow Pipelines (KFP) Provider Service** is included within this project to interface directly with
Kubeflow Pipelines. This service acts as a bridge between the KFP Operator and Vertex AI, facilitating operations such as pipeline
submission, status monitoring, schedules and experiment management.

![KFP Provider]({{< param "subpath" >}}/master/images/kfp.svg)

## Deployment and Usage

Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Kubeflow. [See the getting started guide.](../../../getting-started/installation/#providers)

The configuration can be managed via the [provider custom resource](../../resources/provider/#kubeflow) installed by the operator.

For detailed implementation code and further technical insights, refer to the
[KFP Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/kfp) in the
official repository.

## Implementation Details

- **API** : Implements the [openapi spec for provider services](../overview/#api).

- **Event Handling**: The KFP provider creates run-completion events when reading the status of workflows triggered by
kubeflow. These events are then processed and sent to the operators webhook to update the status of the run.
