---
title: "Kubeflow Pipelines"
---

## Overview

The **Kubeflow Pipelines Provider Service** is included within this project to interface directly with the
Kubeflow Pipelines V2 backend. This service acts as a bridge between the KFP Operator and Kubeflow Pipelines V2,
facilitating operations such as pipeline submission, status monitoring, schedules and experiment management.

![KFP Provider]({{< param "subpath" >}}/master/images/kfp.svg)

## Deployment and Usage

Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Kubeflow. [See the getting started guide.](../../../getting-started/installation/#providers)

Kubeflow Pipelines must be [installed](https://www.kubeflow.org/docs/components/pipelines/operator-guides/installation/).
Its configuration can be controlled using the [KFP specific parameters within a Provider Resource](../../resources/provider/#kubeflow).

For detailed implementation code and further technical insights, refer to the
[KFP Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/kfp) in the
official repository.

## Implementation Details

- **API** : Implements the [openapi spec for provider services](../overview/#api).

- **Event Handling**: The KFP provider creates run-completion events when reading the status of workflows triggered by
Kubeflow. These events are then processed and sent to the operators webhook to update the status of the run.
