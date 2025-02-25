---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name          | Description                                                                                                                                                                                                   | Example                                          |
|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|
| `defaultExperiment`     | Default Experiment name to be used for creating pipeline runs                                                                                                                                                 | `Default`                                        |
| `defaultProvider`       | Default provider name to be used (see [Using Multiple Providers](../providers)). **Note:** This is deprecated as of v1alpha6 and will be removed on release of v1alpha7                                       | `vertex-ai-europe`                               |
| `multiversion`          | If enabled, it will support previous versions of the CRDs, only the latest otherwise                                                                                                                          | `true`                                           |
| `workflowNamespace`     | Namespace where operator Argo workflows should be running - defaults to the operator's namespace                                                                                                              | `kfp-operator-workflows`                         |
| `runCompletionTTL`      | Duration string for how long to keep one-off runs after completion - a zero-length or negative duration will result in runs being deleted immediately after completion; defaults to empty (never delete runs) | `10m`                                            |
| `runCompletionFeed`     | [Configuration of the service](#run-completion-feed-configuration) for the run completion feed back to KFP Operator                                                                                           |                                                  |
| `frameworks`            | Map of pipeline framework to docker image. Pipeline framework to match value in [Pipeline resource](../resources/pipeline)                                                                                    | `default: kfp-operator-argo-kfp-compiler:latest` |
| `defaultProviderValues` | [Configuration of the deployment and service](#provider-values-configuration) created for [providers](../reference/providers/overview)                                                                        |                                                  |


## Run Completion Feed Configuration

| Parameter name | Description                                                                            | Example                                                                                    |
|----------------|----------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| `port`         | The port that the feed endpoint will listen on                                         | `8082`                                                                                     |
| `endpoints`    | Array of run completion event handler endpoints that should be called per feed message | `- host: run-completion-event-handler<br/>&nbsp;&nbsp;path: /<br/>&nbsp;&nbsp;port: 12000` |

## Provider Values Configuration

| Parameter name         | Description                                                                                                                | Example            |
|------------------------|----------------------------------------------------------------------------------------------------------------------------|--------------------|
| `replicas`             | Number of replicas used within the deployment                                                                              | `2`                |
| `serviceContainerName` | Name of the container that will execute the provider image.  **Note:**  this must match the podTemplateSpec name           | `provider-service` |
| `servicePort`          | The port that should expose the service  **Note:**  this must match the podTemplateSpec ports                              | `8080`             |
| `podTemplateSpec`      | [k8s pod template spec for the provider service deployment](https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates) |                    |

An example configuration:
{{% readfile file="/includes/master/reference/controller_manager_config.yaml" code="true" lang="yaml"%}}
