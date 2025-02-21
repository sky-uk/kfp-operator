---
title: "Kubeflow Pipelines"
---

## Overview

The KFP Operator's **Kubeflow Pipelines (KFP) Provider Service** is a specialized component designed to interface 
directly with Kubeflow Pipelines, enabling seamless management and execution of machine learning workflows. This service
acts as a bridge between the KFP Operator and the Kubeflow Pipelines platform, facilitating operations such as pipeline 
submission, status monitoring, and schedules and experiment management.

![KFP Provider]({{< param "subpath" >}}/master/images/kfp.svg)

## Deployment and Usage

To deploy the KFP Provider Service:

### **Configuration**: 
Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Kubeflow Pipelines instance.

For detailed implementation code and further technical insights, refer to the
[KFP Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/kfp) in the
official repository.

## Implementation Details

The implementation of the KFP Provider Service is structured to align with the architecture of the KFP Operator, 
ensuring modularity and ease of integration. Key aspects include:

- **API Client Integration**: Utilizes a client to interact with the Kubeflow Pipelines API, handling authentication, 
request formatting, and response parsing.

- **Event Handling**: Implements mechanisms to process events related to pipeline runs, such as completion notifications 
and error handling, ensuring that the KFP Operator can respond appropriately to changes in run statuses.

- **Configuration Management**: Supports configurable parameters to adapt to different deployment environments and user
requirements, such as API endpoints, authentication tokens, and resource quotas.

