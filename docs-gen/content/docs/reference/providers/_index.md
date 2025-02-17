---
title: "Providers"
weight: 3
---
The KFP operator supports multiple pipeline orchestration providers:
- Vertex AI
- Kubeflow Pipelines

It is also possible to use the KFP Operator with a different provider, but you must provide your own implementation of a [Provider Service](#provider-service).

> Note
Changing the provider of a resource that was previously managed by another provider will result in the resource erroring.
Any referenced resources must always match the provider of the referencing resource (e.g. RunConfiguration to Pipeline) as updates are not propagated or checked and will result in runtime errors on the provider. 


## Provider Service
A provider service acts as the bridge between the KFP Operator and the pipeline orchestration provider. It is responsible for:
- creating the resources on the given provider platform, e.g. Runs in Vertex AI
- reporting the state of any resources on the platform back to the operator

The API must adhere to the OpenAPI spec as below:

{{< swaggerui src="https://raw.githubusercontent.com/sky-uk/kfp-operator/refs/heads/master/provider-service/api/openapi.yaml" >}}


## Provider CLI
A provider CLI acts as ... 
