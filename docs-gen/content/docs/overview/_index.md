---
title: "Overview"
weight: 1
---

The Kubeflow Pipelines Operator provides a declarative API for managing and running machine learning pipelines on Kubeflow with Resource Definitions.

## TFX Pipelines and Components

Unlike imperative Kubeflow Pipelines deployments, the operator takes care of providing all environment-specific configuration and setup for the pipelines. Pipeline creators therefore don't have to provide DAG runners, metadata configs, serving directories, etc. Furthermore, pusher is not required and the operator can extend the pipeline with this very environment-specific component.

For running a pipeline using the operator, only the list of TFX components needs to be returned. Everything else is done by the operator. See the [penguin pipeline]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/quickstart/penguin_pipeline/pipeline.py) for an example.

### Lifecycle phases and Parameter types

TFX Pipelines go through certain lifecycle phases that are unique to this technology. It is helpful to understand where these differ and where they are executed.

**Development:** Creating the components definition as code.

**Compilation:** Applying compile-time parameters and defining the execution runtime (aka DAG runner) for the pipeline to be compiled into a deployable artifact.

**Deployment:** Creating a pipeline representation in the target environment.

**Running:** Instantiating the pipeline, applying runtime parameters and running all pipeline steps involved to completion.

*Note:* Local runners usually skip compilation and deployment and run the pipeline straight away.

TFX allows the parameterisation of Pipelines in most lifecycle stages:

| Parameter type | Description | Example |
| --- | --- | --- |
| Named Constants | Code constants | ANN layer size |
| Compile-time parameter | Parameters that are unlikely to change between pipeline runs supplied as environment variabels to the pipeline function | Bigquery dataset |
| Runtime parameter | Parameters exposed as TFX [RuntimeParameter](https://www.tensorflow.org/tfx/api_docs/python/tfx/v1/dsl/experimental/RuntimeParameter?hl=en) which can be overridden at runtime allow simplified experimentation without having to recompile the pipeline | Number of training runs |

The pipeline operator supports the application of compile time and runtime parameters through its custom resources. We strongly encourage the usage of both of these parameter types to speed up development and experimentation lifecycles. Note that Runtime parameters can initialised to default values from both constants and compile-time parameters

## Eventing Support

The Kubeflow Pipelines operator can optionally be installed with [Argo-Events](https://argoproj.github.io/argo-events/) eventsources which lets users react to events.

Currently, we support the following eventsources:

- [Run Completion Eventsource](../reference/run-completion)

## Architecture Overview

![Architecture](/kfp-operator/architecture.png)

## Limitations

- The operator currently only supports TFX Pipelines - pipelines created using the KFP DSL are not supported yet.
- The operator currently only supports KFP standalone - a full KFP installation is not supported yet.
