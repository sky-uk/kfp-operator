---
title: "Installation"
weight: 3
---

We recommend the installation using Helm as it allows a declarative approach to managing Kubernetes resources.

This guide assumes you are familiar with [Helm](https://helm.sh/).

## Prerequisites

- [Argo 3.1.6-3.3](https://argoproj.github.io/argo-workflows/installation/) installed cluster-wide or into the namespace where the operator's workflows run (see [configuration](../../reference/configuration)).
- [Argo-Events 1.7.4+](https://argoproj.github.io/argo-events/installation/) installed cluster-wide (see [configuration](../../reference/configuration)).

## KFP-Operator

To get a working installation you will need to install both the KFP-Operator and at least one provider ([see below]({{< ref "#providers" >}} "Providers"))

### Build and Install

Create basic `values.yaml` with the following content:

{{% readfile file="/includes/master/quickstart/resources/values.yaml" code="true" lang="yaml" %}}

Install the latest version of the operator

```sh
helm install oci://ghcr.io/kfp-operator/kfp-operator -f values.yaml
```

You will need to configure service accounts and roles required by your chosen `Provider`, [see here for reference]({{< ref "#provider-rbac" >}} "Provider RBAC Reference").

### Configuration Values

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
| `manager.multiversion.storedVersion`                      | Specifies which CRD version should be set as the stored version. Only takes effect if `manager.multiversion.enabled` is set to `true`. Defaults to the latest version.                                              |
| `manager.webhookCertificates.provider`                    | K8s conversion webhook TLS certificate provider - choose `cert-manager` for Helm to deploy certificates if cert-manager is available or `custom` otherwise (see below)                                              |
| `manager.webhookCertificates.secretName`                  | Name of a K8s secret deployed into the operator namespace to secure the webhook endpoint with, required if the `custom` provider is chosen                                                                          |
| `manager.webhookCertificates.caBundle`                    | CA bundle of the certificate authority that has signed the webhook's certificate, required if the `custom` provider is chosen                                                                                       |
| `manager.webhookServicePort`                              | Port for the webhook service to listen on - defaults to 9443                                                                                                                                                        |
| `manager.runcompletionWebhook.servicePort`                | Port for the run completion event webhook service to listen on - defaults to 8082                                                                                                                                   |
| `manager.runcompletionWebhook.endpoints`                  | Array of endpoints for the run completion event handlers to be called when a run completion event is passed                                                                                                         |
| `manager.pipeline.frameworks`                             | Map of additional pipeline frameworks to their respective container images - defaults to empty                                                                                                                      |
| `logging.verbosity`                                       | Logging verbosity for all components - see the [logging documentation]({{< param "github_project_repo" >}}/blob/master/CONTRIBUTING.md#logging) for valid values                                                    |
| `statusFeedback.enabled`                                  | Whether run completion eventing and status update feedback loop should be installed - defaults to `false`                                                                                                           |

Examples for these values can be found in the [test configuration]({{< ghblob "/helm/kfp-operator/test/values.yaml" >}})

## Providers

Please refer to your chosen provider instructions before proceeding. Supported providers are:
- [Vertex AI](../../reference/providers/vai/#deployment-and-usage)

To install your chosen provider, create a [Provider resource](../../reference/resources/provider) in a namespace that the operator can access (see the [rbac setup below]({{< ref "#provider-rbac" >}}) for reference). Once it is applied the Provider controller will reconcile and create the Provider Deployment and Provider Service within the same namespace that the Provider resource was applied.

## Role-based access control (RBAC) for providers {#provider-rbac}
When using a provider, you should create the necessary `ServiceAccount`, `RoleBinding` and `ClusterRoleBinding` resources required for the providers being used.

In order for Event Source Servers and the Controller to read the Providers you must configure their service accounts
to have read permissions of Provider resources. e.g:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfp-operator-kfp-providers-viewer-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kfp-operator-providers-viewer-role
subjects:
- kind: ServiceAccount
  name: kfp-operator-kfp #Used by Event Source Server
  namespace: kfp-operator-system
- kind: ServiceAccount
  name: kfp-operator-controller-manager #Used by KFP Controller
  namespace: kfp-operator-system
```

An example configuration for Providers is also provided below for reference:
```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kfp-operator-kfp-service-account
  namespace: kfp-namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfp-operator-kfp-runconfiguration-viewer-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kfp-operator-runconfiguration-viewer-role
subjects:
- kind: ServiceAccount
  name: kfp-operator-kfp-service-account
  namespace: kfp-namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfp-operator-kfp-run-viewer-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kfp-operator-run-viewer-role
subjects:
- kind: ServiceAccount
  name: kfp-operator-kfp-service-account
  namespace: kfp-namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kfp-operator-provider-workflow-executor
  namespace: kfp-namespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kfp-operator-workflow-executor
subjects:
- kind: ServiceAccount
  name: kfp-operator-kfp-service-account
  namespace: kfp-namespace
```
