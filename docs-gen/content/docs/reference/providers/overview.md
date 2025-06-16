---
title: "Overview"
type: swagger
weight: 1
---

The KFP Operator supports the following pipeline orchestration providers:
- **[Vertex AI Integration](../vai)**

You can also integrate the KFP Operator with custom providers by implementing a [custom Provider Service](#using-custom-providers).

## Service

A provider service bridges the KFP Operator and the pipeline orchestration provider. It performs key tasks such as:

- **Eventing**: Reports the state of resources on the provider to the KFP Operator.
- **Resource Management**: Manages provider-specific resources, such as runs in Vertex AI.

The KFP Operator will deploy the Provider service as Kubernetes deployment with an accompanying Kubernetes Service based
off the [configuration provided.](#configuration)

![provider-controller]({{< param "subpath" >}}/master/images/provider-controller.svg)

Interaction with this service is via argo-workflows, whereby http requests are sent to the service to perform actions on the provider.

### Eventing

The provider service is the first point of contact for [Run Completion Events](../../run-completion) received from the 
external pipeline orchestration provider. In its current implementation, the KFP Operator supports:

- **Vertex AI**: Run completion events are consumed from Pub/Sub.

For each provider, the events are processed to ensure accurate status reporting back to the KFP Operator.

### API

The management of resources for each provider can be handled through an HTTP API. Custom providers can be integrated by 
implementing a service that adheres to the OpenAPI specification.

{{< spoiler "View the OpenAPI Specification" >}}
{{< swaggerui src="master/openapi.yaml" >}}
{{< /spoiler >}}

The specification outlines the structure, endpoints, and methods required for full integration with the KFP Operator.

### Configuration
Configuration of a provider service is managed through 2 separate configuration components: 

- #### Provider custom resource 

  The provider custom resource is designed to provide the configuration for how the provider should behave, ie, what provider it should use, the cli required including the compilation
  and how to interact with the provider service. See the [provider custom resource](../../resources/provider) for more information.

- #### Operator configuration. 

  The operator configuration is designed to provide the configuration for the underlying provider service deployment, ie, the ports to expose, the cpu / memory allocation.
  For more information see the [operator configuration](../../configuration).

## Using Custom Providers

To use a custom provider:

1. **Implement a Provider Service**: Ensure the service adheres to the OpenAPI specification and handles eventing and state reporting appropriately.
2. **Configure Run Completion Events**: Integrate your provider with an eventing mechanism compatible with the KFP Operator. See the [Run Completion Events](../../run-completion) documentation for more information.
3. **Deploy and Test**: Deploy the custom provider and verify proper communication with the KFP Operator.

