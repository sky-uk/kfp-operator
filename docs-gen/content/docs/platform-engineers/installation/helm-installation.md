---
title: "Helm Installation Guide"
linkTitle: "Helm Installation"
description: "Complete guide to install the KFP Operator using Helm with production-ready configuration"
weight: 10
---

# Installing the KFP Operator

This guide provides comprehensive instructions for installing the Kubeflow Pipelines Operator in your Kubernetes cluster. We recommend using Helm for a declarative approach to managing Kubernetes resources.

## Overview

The KFP Operator installation consists of three main components:

1. **Prerequisites**: Required dependencies (Argo Workflows, Argo Events)
2. **KFP Operator**: The core operator and controllers
3. **Providers**: ML orchestration platform integrations (KFP, Vertex AI)

## Prerequisites

Before installing the KFP Operator, ensure your cluster meets these requirements:

### Cluster Requirements

- **Kubernetes**: v1.21 or later (tested up to v1.28)
- **Cluster Admin Access**: Required for installing CRDs and cluster-wide resources
- **Storage**: Persistent storage for pipeline artifacts and metadata
- **Network**: Outbound internet access for downloading container images

### Required Dependencies

#### 1. Argo Workflows
**Version**: 3.1.6 - 3.4.x
**Purpose**: Workflow execution engine for pipeline orchestration

```bash
# Install Argo Workflows (cluster-wide)
kubectl create namespace argo
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.4/install.yaml
```

**Verification:**
```bash
kubectl get pods -n argo
# Should show argo-server and workflow-controller pods running
```

#### 2. Argo Events (Optional but Recommended)
**Version**: 1.7.4 or later
**Purpose**: Event-driven pipeline automation

```bash
# Install Argo Events (cluster-wide)
kubectl create namespace argo-events
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-events/stable/manifests/install.yaml
```

**Verification:**
```bash
kubectl get pods -n argo-events
# Should show eventbus-controller and eventsource-controller pods running
```

### Optional Dependencies

#### Cert-Manager (Recommended for Production)
**Purpose**: Automatic TLS certificate management for webhooks

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

#### Prometheus Operator (For Monitoring)
**Purpose**: Metrics collection and monitoring

```bash
# Install Prometheus Operator
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

### Helm Installation

This guide assumes you have [Helm 3.x](https://helm.sh/docs/intro/install/) installed:

```bash
# Verify Helm installation
helm version
# Should show version 3.x
```

## Installing the KFP Operator

The KFP Operator installation consists of the core operator and at least one provider for ML orchestration.

### Quick Installation

For a basic installation with default settings:

```bash
# Add the KFP Operator Helm repository
helm repo add kfp-operator https://sky-uk.github.io/kfp-operator/
helm repo update

# Install with default values
helm install kfp-operator kfp-operator/kfp-operator
```

### Custom Installation

For production deployments, create a custom `values.yaml` file:

```yaml
# values.yaml - Basic configuration
namespace:
  create: true
  name: kfp-operator-system

manager:
  replicas: 2  # High availability
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

  # Argo Workflows configuration
  argo:
    serviceAccount:
      create: true
      name: kfp-operator-argo
    stepTimeoutSeconds:
      default: 600  # 10 minutes
      compile: 3600  # 1 hour
    ttlStrategy:
      secondsAfterCompletion: 3600  # Clean up after 1 hour

  # Monitoring configuration
  monitoring:
    create: true
    serviceMonitor:
      create: true  # For Prometheus Operator

# Enable event-driven workflows
statusFeedback:
  enabled: true

# Logging configuration
logging:
  verbosity: 1  # 0=error, 1=info, 2=debug
```

Install with custom configuration:

```bash
helm install kfp-operator kfp-operator/kfp-operator -f values.yaml
```

### Installation Methods

#### Method 1: Helm Repository (Recommended)
```bash
# Add repository and install
helm repo add kfp-operator https://sky-uk.github.io/kfp-operator/
helm install kfp-operator kfp-operator/kfp-operator -f values.yaml
```

#### Method 3: Local Development (Requires local kubernetes cluster)
```bash
# Clone repository and install from source
git clone https://github.com/sky-uk/kfp-operator.git
cd kfp-operator
helm install kfp-operator ./helm/kfp-operator -f values.yaml
```

### Verification

Verify the installation was successful:

```bash
# Check operator pods
kubectl get pods -n kfp-operator-system

# Check CRDs were installed
kubectl get crd | grep pipelines.kubeflow.org

# Check operator logs
kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager
```

Expected output:
```
NAME                                           READY   STATUS    RESTARTS   AGE
kfp-operator-controller-manager-xxx-xxx        2/2     Running   0          2m
```

### Namespace Configuration

The operator can be installed in any namespace. Common patterns:

#### Dedicated Namespace (Recommended)
```yaml
namespace:
  create: true
  name: kfp-operator-system
```

#### Existing Namespace
```yaml
namespace:
  create: false
  name: ml-platform
```

#### Multi-tenant Setup
```yaml
# Install operator in system namespace
namespace:
  name: kfp-operator-system

# Configure RBAC for multiple tenant namespaces
manager:
  rbac:
    create: true
    # Additional cluster roles will be created
```

### Configuration Values

Valid configuration options to override the [Default `values.yaml`]({{< ghblob "/helm/kfp-operator/values.yaml" >}}) are:

| Parameter name                                            | Description                                                                                                                                                                                                         |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `containerRegistry`                                       | Container Registry base path for all container images                                                                                                                                                               |
| `namespace.create`                                        | Create the namespace for the operator                                                                                                                                                                               |
| `namespace.name`                                          | Operator namespace name                                                                                                                                                                                             |
| `manager.argo.containerDefaults`                          | Container Spec defaults to be used for Argo workflow pods created by the operator                                                                                                                                   |
| `manager.argo.metadata`                                   | Container Metadata defaults to be used for Argo workflow pods created by the operator                                                                                                                               |
| `manager.argo.securityContext`                            | [Security Context](https://argo-workflows.readthedocs.io/en/latest/workflow-pod-security-context/) applied to Argo WorkflowTemplate. To run as root user, set `securityContext` to `null` or `securityContext.runAsNonRoot` to `false` |
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
| `manager.leaderElection.enabled`                          | Toggle leader election - defaults to `true`                                                                                                                                                                         |
| `manager.leaderElection.id`                               | Leader election Lease resource name - defaults to `kfp-operator-lock`                                                                                                                                               |
| `manager.resources`                                       | Manager resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources)                                                                              |
| `manager.configuration`                                   | Manager configuration as defined in [Configuration](../configuration/operator-configuration) (note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied) |
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
- [Kubeflow Pipelines V2](../configuration/providers/kfp/#deployment-and-usage)
- [Vertex AI](../configuration/providers/vai/#deployment-and-usage)

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

## Kubeflow completion eventing required RBACs
If using the `Kubeflow Pipelines` Provider you will also need a `ClusterRole` for permission to interact with argo workflows for the
[eventing system](../events/run-completion/) for run completion events.


```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kfp-operator-kfp-eventsource-server-role
rules:
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - get
  - list
  - patch
  - update
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfp-operator-kfp-eventsource-server-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kfp-operator-kfp-eventsource-server-role
subjects:
- kind: ServiceAccount
  name:  kfp-operator-kfp-service-account
  namespace:  kfp-operator-namespace
```
