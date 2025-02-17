---
title: "Installation"
weight: 2
---

We recommend the installation using Helm as it allows a declarative approach to managing Kubernetes resources.

This guide assumes you are familiar with [Helm](https://helm.sh/).

## Prerequisites

- [Argo 3.1.6-3.3](https://argoproj.github.io/argo-workflows/installation/) installed cluster-wide or into the namespace where the operator's workflows run (see [configuration](../../reference/configuration)).
- [Argo-Events 1.7.4+](https://argoproj.github.io/argo-events/installation/) installed cluster-wide (see [configuration](../../reference/configuration)).
- The KFP-Operator supports configurable provider backends. Currently, Kubeflow Pipelines and Vertex AI are supported. Please refer to the [respective configuration section](../../reference/configuration/#provider-configuration) before proceeding.

## Build and Install

Create basic `values.yaml` with the following content:

{{% readfile file="/includes/archive/v0.6.0/quickstart/resources/values.yaml" code="true" lang="yaml" %}}

Install the latest version of the operator

```sh
helm install oci://ghcr.io/kfp-operator/kfp-operator -f values.yaml
```

## Configuration Values

Valid configuration options to override the [Default `values.yaml`]({{< ghblob "/helm/kfp-operator/values.yaml" >}}) are:

| Parameter name                                            | Description                                                                                                                                                                                                         |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `containerRegistry`                                       | Container Registry base path for all container images                                                                                                                                                               |
| `namespace.create`                                        | Create the namespace for the operator                                                                                                                                                                               |
| `namespace.name`                                          | Operator namespace name                                                                                                                                                                                             |
| `manager.argo.containerDefaults`                          | Container Spec defaults to be used for Argo workflow pods created by the operator                                                                                                                                   |
| `manager.argo.metadata`                                   | Container Metadata defaults to be used for Argo workflow pods created by the operator                                                                                                                               |
| `manager.argo.ttlStrategy`                                | [TTL Strategy](https://argoproj.github.io/argo-workflows/fields/#ttlstrategy) used for all Argo Workflows                                                                                                           |
| `manager.argo.stepTimeoutSeconds.compile`                 | Timeout in seconds for compiler steps - defaults to 1800 (30m)                                                                                                                                                      |
| `manager.argo.stepTimeoutSeconds.default`                 | Default [timeout in seconds](https://argoproj.github.io/argo-workflows/walk-through/timeouts/) for workflow steps - defaults to 300 (5m)                                                                            |
| `manager.argo.serviceAccount.name`                        | The [k8s service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) used to run Argo workflows                                                                           |
| `manager.argo.serviceAccount.create`                      | Create the Argo Workflows service account (or assume it has been created externally)                                                                                                                                |
| `manager.argo.serviceAccount.metadata`                    | Optional Argo Workflows service account default metadata                                                                                                                                                            |
| `manager.metadata`                                        | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the manager's pods                                                                            |
| `manager.rbac.create`                                     | Create roles and rolebindings for the operator                                                                                                                                                                      |
| `manager.serviceAccount.name`                             | Manager service account's name                                                                                                                                                                                      |
| `manager.serviceAccount.create`                           | Create the manager's service account or expect it to be created externally                                                                                                                                          |
| `manager.replicas`                                        | Number of replicas for the manager deployment                                                                                                                                                                       |
| `manager.resources`                                       | Manager resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources)                                                                              |
| `manager.configuration`                                   | Manager configuration as defined in [Configuration](../../reference/configuration) (note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied) |
| `manager.monitoring.create`                               | Create the manager's monitoring resources                                                                                                                                                                           |
| `manager.monitoring.rbacSecured`                          | Enable addtional RBAC-based security                                                                                                                                                                                |
| `manager.monitoring.serviceMonitor.create`                | Create a ServiceMonitor for the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)                                                                                                   |
| `manager.monitoring.serviceMonitor.endpointConfiguration` | Additional configuration to be used in the service monitor endpoint (path, port and scheme are provided)                                                                                                            |
| `manager.multiversion.enabled`                            | Enable multiversion API. Should be used in production to allow version migration, disable for simplified installation                                                                                               |
| `manager.webhookCertificates.provider`                    | K8s conversion webhook TLS certificate provider - choose `cert-manager` for Helm to deploy certificates if cert-manager is available or `custom` otherwise (see below)                                              |
| `manager.webhookCertificates.secretName`                  | Name of a K8s secret deployed into the operator namespace to secure the webhook endpoint with, required if the `custom` provider is chosen                                                                          |
| `manager.webhookCertificates.caBundle`                    | CA bundle of the certificate authority that has signed the webhook's certificate, required if the `custom` provider is chosen                                                                                       |
| `manager.provider.type`                                   | Provider type (`kfp` for Kubeflow Pipelines or `vai` for Vertex AI Pipelines)                                                                                                                                       |
| `manager.provider.configuration`                          | Configuration block for the specific provider (see [Provider Configuration](../../reference/configuration#provider-configuration)), automatically mounted as a file                                                 |
| `logging.verbosity`                                       | Logging verbosity for all components - see the [logging documentation]({{< param "github_project_repo" >}}/blob/master/CONTRIBUTING.md#logging) for valid values                                                    |
| `eventsourceServer.metadata`                              | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the eventsource server's pods                                                                 |
| `eventsourceServer.rbac.create`                           | Create roles and rolebindings for the eventsource server                                                                                                                                                            |
| `eventsourceServer.serviceAccount.name`                   | Eventsource server's service account                                                                                                                                                                                |
| `eventsourceServer.serviceAccount.create`                 | Create the eventsource server's service account or expect it to be created externally                                                                                                                               |
| `eventsourceServer.resources`                             | Eventsource server resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources)                                                                   |
| `providers`                                               | Dictionary of providers (see below)                                                                                                                                                                                 |

Examples for these values can be found in the [test configuration]({{< ghblob "/helm/kfp-operator/test/values.yaml" >}})

### Providers

The `providers` block contains a dictionary of provider names to provider configurations:

| Parameter name            | Description                                                                                                                                                 |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`                    | Provider type (`kfp` or `vai`)                                                                                                                              |
| `serviceAccount.name`     | Name of the service account to run provider-specific operations                                                                                             |
| `serviceAccount.create`   | Create the service account (or assume it has been created externally)                                                                                       |
| `serviceAccount.metadata` | Optional service account default metadata                                                                                                                   |
| `configuration`           | See [Provider Configuration](../../reference/configuration/#provider-configurations) for all available providers and their respective configuration options |

Example:

```yaml
providers:
  kfp:
    type: kfp
    serviceAccount:
      name: kfp-operator-kfp
      create: false
    configuration:
      ...
  vai:
    type: vai
    serviceAccount: 
      name: kfp-operator-kfp
      create: true
      metadata:
        annotations:
          iam.gke.io/gcp-service-account: kfp-operator-vai@my-project.iam.gserviceaccount.com
    configuration:
      ...
```
