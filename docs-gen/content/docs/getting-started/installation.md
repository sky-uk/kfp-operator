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

| Parameter name | Description |
| --- | --- |
| `containerRegistry` | Container Registry base path for all container images |
| `namespace.create` | Create the namespace for the operator |
| `namespace.name` | Operator namespace name |
| `manager.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the manager's pods |
| `manager.rbac.create` | Create roles and rolebindings for the operator |
| `manager.serviceAccount.create` | Create the manager's service account |
| `manager.serviceAccount.name` | Manager service account's name |
| `manager.replicas` | Number of replicas for the manager deployment |
| `manager.resources` | Manager resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources) |
| `manager.configuration` | Manager configuration as defined in [Configuration](../../reference/configuration) (note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied) |
| `manager.monitoring.create` | Create the manager's monitoring resources |
| `manager.monitoring.rbacSecured` | Enable addtional RBAC-based security |
| `manager.monitoring.serviceMonitor.create` | Create a ServiceMonitor for the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) |
| `manager.monitoring.serviceMonitor.endpointConfiguration` | Additional configuration to be used in the service monitor endpoint (path, port and scheme are provided) |
| `logging.verbosity` | Logging verbosity for all components. See the [logging documentation]({{< param "github_project_repo" >}}/blob/master/CONTRIBUTING.md#logging) for valid values |
| `eventsourceServer.create` | Create the [Argo-Events eventsource server](../../reference/run-completion) |
| `eventsourceServer.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the eventsource server's pods |
| `eventsourceServer.port` | Service port of the eventsource server |
| `eventsourceServer.rbac.create` | Create roles and rolebindings for the eventsource server |
| `eventsourceServer.serviceAccount.create` | Create the eventsource server's service account |
| `eventsourceServer.serviceAccount.name` | Eventsource server's service account |
| `eventsourceServer.resources` | Eventsource server resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources) |

Examples for these values can be found in the [test configuration]({{< ghblob "/helm/kfp-operator/test/values.yaml" >}})
