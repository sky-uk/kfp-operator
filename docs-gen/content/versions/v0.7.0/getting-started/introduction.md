---
title: "Introduction"
weight: 2
---

## Compatibility

The operator currently supports
- TFX Pipelines with Python 3.7 and 3.9 - pipelines created using the KFP DSL are not supported yet
- KFP standalone (a full KFP installation is not supported yet) and Vertex AI

## TFX Pipelines and Components

Unlike imperative Kubeflow Pipelines deployments, the operator takes care of providing all environment-specific configuration and setup for the pipelines. Pipeline creators therefore don't have to provide DAG runners, metadata configs, serving directories, etc. Furthermore, pusher is not required and the operator can extend the pipeline with this very environment-specific component.

For running a pipeline using the operator, only the list of TFX components needs to be returned. Everything else is done by the operator. See the [penguin pipeline]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/versions/v0.7.0/quickstart/penguin_pipeline/pipeline.py) for an example.

### Lifecycle phases and Parameter types

TFX Pipelines go through certain lifecycle phases that are unique to this technology. It is helpful to understand where these differ and where they are executed.

**Development:** Creating the components definition as code.

**Compilation:** Applying compile-time parameters and defining the execution runtime (aka DAG runner) for the pipeline to be compiled into a deployable artifact.

**Deployment:** Creating a pipeline representation in the target environment.

**Running:** Instantiating the pipeline, applying runtime parameters and running all pipeline steps involved to completion.

*Note:* Local runners usually skip compilation and deployment and run the pipeline straight away.

TFX allows the parameterization of Pipelines in most lifecycle stages:

| Parameter type         | Description                                                                                                                                                                                                                                              | Example                 |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| Named Constants        | Code constants                                                                                                                                                                                                                                           | ANN layer size          |
| Compile-time parameter | Parameters that are unlikely to change between pipeline runs supplied as environment variabels to the pipeline function                                                                                                                                  | Bigquery dataset        |
| Runtime parameter      | Parameters exposed as TFX [RuntimeParameter](https://www.tensorflow.org/tfx/api_docs/python/tfx/v1/dsl/experimental/RuntimeParameter?hl=en) which can be overridden at runtime allow simplified experimentation without having to recompile the pipeline | Number of training runs |

The pipeline operator supports the application of compile time and runtime parameters through its custom resources. We strongly encourage the usage of both of these parameter types to speed up development and experimentation lifecycles. Note that Runtime parameters can initialised to default values from both constants and compile-time parameters

## Eventing Support

The Kubeflow Pipelines operator can optionally be installed with [Argo-Events](https://argoproj.github.io/argo-events/) eventsources which lets users react to events.

Currently, we support the following eventsources:

- [Run Completion Eventsource](../../reference/run-completion)

## Architecture Overview

The KFP Operator follows [the standard Kubernetes operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), where a *controller* manages the state of each [custom resource](../../reference/resources/). Each controller creates [Argo Workflows](https://argoproj.github.io/workflows/) that make calls to the [Provider Service](../../reference/providers/overview), which in turn call the Orchestration Provider's API (e.g. Vertex AI).

The KFP Operator also handles Run Completion Events, extracted from the Orchestration Provider, and publishes these events to an [EventBus](https://argoproj.github.io/argo-events/eventbus/eventbus/) for clients to react to.

![Architecture]({{< param "subpath" >}}/versions/v0.7.0/images/architecture.svg)

The sequence of operations that the KFP Operator handles can be roughly broken down into three separate journeys:
- **User Journey**: How the operator reacts when a user submits a custom resource, e.g. a [Pipeline](../../reference/resources/pipeline) or [RunConfiguration](../../reference/resources/runconfiguration)
- **Event Journey**: How the operator reacts to [run completion events](../../reference/run-completion)
- **Provider Management**: How the operator manages the state of [the Provider custom resource](../../reference/resources/provider)


![KFP Operator sequence diagram]({{< param "subpath" >}}/versions/v0.7.0/images/sequence-diagram.svg)
