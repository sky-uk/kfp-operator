---
title: "Overview"
weight: 1
---

The KFP Operator supports multiple pipeline orchestration providers, currently including:
- **[Vertex AI Integration](../providers/vai)**
- **[Kubeflow Pipelines Integration](../providers/kfp)**

You can also integrate the KFP Operator with custom providers by implementing a [custom Provider Service](#using-custom-providers).


## Service

A provider service bridges the KFP Operator and the pipeline orchestration provider. It performs key tasks such as:

- **State Reporting**: Reports the state of resources on the provider to the KFP Operator.
- **Resource Creation**: Creates provider-specific resources, such as runs in Vertex AI.

The KFP Operator will deploy the Provider service as Kubernetes deployment with an accompanying Kubernetes Service based off
the configuration provided in the provider custom resource. 

![provider-controller]({{< param "subpath" >}}/master/images/provider-controller.svg)

### Eventing

The provider service is the first point of contact for [Run Completion Events](../../run-completion) received from the external pipeline orchestration provider. In its current implementation, the KFP Operator supports:

- **Vertex AI**: Run completion events are consumed from Pub/Sub.
- **Kubeflow Pipelines**: Run completion events are consumed from workflows via Argo Events.

For each provider, the events are processed to ensure accurate status reporting back to the KFP Operator.

### API

The management of resources for each provider can be handled through an HTTP API. Custom providers can be integrated by implementing a service that adheres to the OpenAPI specification.

{{< spoiler "Viewing the OpenAPI Specification" >}}
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



## CLI

The provider CLI image facilitates interaction with the provider service via its API, primarily from Argo workflows. The CLI is responsible for:

- **Model Compilation**: Compiles models into manifests and submits them to the provider service.
- **Resource Management**: Handles the creation, deletion, and updating of resources by sending requests to the provider service endpoints.

## Using Custom Providers

To use a custom provider:

1. **Implement a Provider Service**: Ensure the service adheres to the OpenAPI specification and handles eventing and state reporting appropriately.
2. **Configure Run Completion Events**: Integrate your provider with an eventing mechanism compatible with the KFP Operator.
3. **Deploy and Test**: Deploy the custom provider and verify proper communication with the KFP Operator.

## Additional Resources

- [Run Completion Events](../run-completion)
- [Vertex AI Integration](../providers/vai)
- [Kubeflow Pipelines Integration](../providers/kfp)

With proper configuration, the KFP Operator can streamline the orchestration and monitoring of pipeline resources across multiple providers or custom implementations.

