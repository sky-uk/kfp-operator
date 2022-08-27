---
title: "Installation"
weight: 2
---

We recommend the installation using Helm as it allows a declarative approach to managing Kubernetes resources.

This guide assumes you are familiar with [Helm](https://helm.sh/).

## Build and Install

At the moment, you will have to build and publish the container images to run the operator manually.
We are looking to publish images to a public repository in the near future.
Please follow the [Development Guide](https://github.com/sky-uk/kfp-operator/blob/master/DEVELOPMENT.md#building-and-publishing) to publish these images.

```bash
helm install -f values.yaml kfp-operator <YOUR_CHART_REPOSITORY>/kfp-operator
```

## Configuration Values

Valid configuration options to override the [Default `values.yaml`]({{< ghblob "/helm/kfp-operator/values.yaml" >}}) are:

| Parameter name                                            | Description                                                                                                                                                                                                         |
|-----------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `containerRegistry`                                       | Container Registry base path for all container images                                                                                                                                                               |
| `namespace.create`                                        | Create the namespace for the operator                                                                                                                                                                               |
| `namespace.name`                                          | Operator namespace name                                                                                                                                                                                             |
| `manager.argo.containerDefaults`                          | Container Spec defaults to be used for Argo workflow pods created by the operator                                                                                                                                   |
| `manager.argo.metadataDefaults`                           | Container Metadata defaults to be used for Argo workflow pods created by the operator                                                                                                                               |
| `manager.argo.serviceAccount`                             | The [k8s Service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) used to run Argo workflows                                                                           |
| `manager.metadata`                                        | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the manager's pods                                                                            |
| `manager.rbac.create`                                     | Create roles and rolebindings for the operator                                                                                                                                                                      |
| `manager.serviceAccount.create`                           | Create the manager's service account                                                                                                                                                                                |
| `manager.serviceAccount.name`                             | Manager service account's name                                                                                                                                                                                      |
| `manager.replicas`                                        | Number of replicas for the manager deployment                                                                                                                                                                       |
| `manager.resources`                                       | Manager resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources)                                                                              |
| `manager.configuration`                                   | Manager configuration as defined in [Configuration](../../reference/configuration) (note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied) |
| `manager.monitoring.create`                               | Create the manager's monitoring resources                                                                                                                                                                           |
| `manager.monitoring.rbacSecured`                          | Enable addtional RBAC-based security                                                                                                                                                                                |
| `manager.monitoring.serviceMonitor.create`                | Create a ServiceMonitor for the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)                                                                                                   |
| `manager.monitoring.serviceMonitor.endpointConfiguration` | Additional configuration to be used in the service monitor endpoint (path, port and scheme are provided)                                                                                                            |
| `manager.multiversion.enabled`                            | Enable multiversion API. Should be used in production to allow version migration. Disable for simplified installation                                                                                               |
| `manager.multiversion.webhookCertificates.provider`       | K8s conversion webhook TLS certificate provider. Choose `cert-manager` for helm to deploy certificates if cert-manager is available. Choose `custom` otherwise                                                      |
| `manager.multiversion.webhookCertificates.secretName`     | Name of a K8s secret deployed into the operator namespace to secure the webhook endpoint with. Required if the `custom` provider is chosen                                                                          |
| `manager.multiversion.webhookCertificates.caBundle`       | CA bundle of the certificate authority that has signed the webhook's certificate. Required if the `custom` provider is chosen                                                                                       |
| `logging.verbosity`                                       | Logging verbosity for all components. See the [logging documentation]({{< param "github_project_repo" >}}/blob/master/CONTRIBUTING.md#logging) for valid values                                                     |
| `eventsourceServer.create`                                | Create the [Argo-Events eventsource server](../../reference/run-completion)                                                                                                                                         |
| `eventsourceServer.metadata`                              | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the eventsource server's pods                                                                 |
| `eventsourceServer.port`                                  | Service port of the eventsource server                                                                                                                                                                              |
| `eventsourceServer.rbac.create`                           | Create roles and rolebindings for the eventsource server                                                                                                                                                            |
| `eventsourceServer.serviceAccount.create`                 | Create the eventsource server's service account                                                                                                                                                                     |
| `eventsourceServer.serviceAccount.name`                   | Eventsource server's service account                                                                                                                                                                                |
| `eventsourceServer.resources`                             | Eventsource server resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources)                                                                   |

Examples for these values can be found in the [test configuration]({{< ghblob "/helm/kfp-operator/test/values.yaml" >}})

## Additional Notes

If you don't have a cluster-wide installation of Argo (for example if you are using the namespaced installation provided by Kubeflow), you will need to apply additional permissions to the Argo service account to read `ClusterWorkflowTemplate`s as described by `argo-clusterworkflowtemplate-role` and `argo-clusterworkflowtemplate-role-binding` in https://github.com/argoproj/argo-workflows/blob/master/manifests/quick-start/base/cluster-workflow-template-rbac.yaml, changing the namespace to match the one where Argo is installed.  
