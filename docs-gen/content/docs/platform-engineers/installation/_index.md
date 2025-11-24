---
title: "Installation"
linkTitle: "Installation"
description: "Complete guide to install and configure the KFP Operator in your Kubernetes cluster"
weight: 10
---

# Installing the KFP Operator

Install and configure the Kubeflow Pipelines Operator in your Kubernetes cluster.

## Overview

The installation includes:

1. **Prerequisites**: Argo Workflows, Argo Events
2. **KFP Operator**: Core operator and controllers
3. **Providers**: ML platform integrations (KFP, Vertex AI)

## What You'll Set Up

- KFP Operator with production-ready configuration
- Provider connections to ML orchestration platforms
- Monitoring and observability
- Security and RBAC policies
- Installation verification with test pipelines

## Prerequisites

### Cluster Requirements

Ensure your cluster meets these requirements:

#### Kubernetes Version
- **Minimum**: Kubernetes v1.21
- **Recommended**: Kubernetes v1.24+
- **Tested**: Up to Kubernetes v1.28

#### Network Requirements
- **Outbound Internet**: For downloading container images
- **Internal DNS**: Kubernetes DNS resolution
- **Load Balancer**: For external access (optional)

### Required Dependencies

#### 1. Argo Workflows
**Version**: 3.1.6 - 3.4.x
**Purpose**: Workflow execution engine for pipeline orchestration

```bash
# Install Argo Workflows (cluster-wide)
kubectl create namespace argo
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.4/install.yaml

# Verify installation
kubectl get pods -n argo
# Should show argo-server and workflow-controller pods running
```

**Production Configuration:**
```yaml
# argo-workflows-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: workflow-controller-configmap
  namespace: argo
data:
  config: |
    # Increase parallelism for production
    parallelism: 100
    # Set resource limits
    resourceRateLimit:
      limit: 10
      burst: 1
    # Configure artifact repository
    artifactRepository:
      s3:
        bucket: my-argo-artifacts
        endpoint: s3.amazonaws.com
        region: us-west-2
```

#### 2. Argo Events (Optional but Recommended)
**Version**: 1.7.4 or later  
**Purpose**: Event-driven pipeline automation

```bash
# Install Argo Events (cluster-wide)
kubectl create namespace argo-events
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-events/stable/manifests/install.yaml

# Verify installation
kubectl get pods -n argo-events
# Should show eventbus-controller and eventsource-controller pods running
```

### Optional Dependencies

#### Cert-Manager (Recommended for Production)
**Purpose**: Automatic TLS certificate management for webhooks

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Verify installation
kubectl get pods -n cert-manager
```

#### Prometheus Operator (For Monitoring)
**Purpose**: Metrics collection and monitoring

```bash
# Install Prometheus Operator
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

## Installing the KFP Operator

### Method 1: Helm Installation (Recommended)

#### Add the Helm Repository

```bash
# Add the KFP Operator Helm repository
helm repo add kfp-operator https://sky-uk.github.io/kfp-operator/
helm repo update

# Verify the repository
helm search repo kfp-operator
```

#### Create Custom Configuration

Create a `values.yaml` file for your environment:

```yaml
# values.yaml - Production configuration
namespace:
  create: true
  name: kfp-operator-system

manager:
  # High availability setup
  replicas: 2
  
  # Resource allocation
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  
  # Argo Workflows configuration
  argo:
    serviceAccount:
      create: true
      name: kfp-operator-argo
    stepTimeoutSeconds:
      default: 1800    # 30 minutes
      compile: 3600    # 1 hour
    ttlStrategy:
      secondsAfterCompletion: 3600  # Clean up after 1 hour
    containerDefaults:
      resources:
        requests:
          cpu: 100m
          memory: 128Mi
        limits:
          cpu: 500m
          memory: 512Mi
  
  # Security configuration
  rbac:
    create: true
  
  # Monitoring configuration
  monitoring:
    create: true
    serviceMonitor:
      create: true  # For Prometheus Operator
  
  # Webhook configuration (production)
  multiversion:
    enabled: true
  webhookCertificates:
    provider: cert-manager  # Use cert-manager for TLS

# Enable event-driven workflows
statusFeedback:
  enabled: true

# Logging configuration
logging:
  verbosity: 1  # 0=error, 1=info, 2=debug

# Container registry configuration
containerRegistry: "your-registry.com/kfp-operator"
```

#### Install the Operator

```bash
# Install with custom configuration
helm install kfp-operator kfp-operator/kfp-operator \
  --namespace kfp-operator-system \
  --create-namespace \
  --values values.yaml

# Verify installation
helm list -n kfp-operator-system
```

### Method 2: OCI Registry Installation

```bash
# Install directly from OCI registry
helm install kfp-operator \
  oci://ghcr.io/sky-uk/kfp-operator/helm/kfp-operator \
  --namespace kfp-operator-system \
  --create-namespace \
  --values values.yaml
```

### Method 3: Development Installation

```bash
# Clone repository and install from source
git clone https://github.com/sky-uk/kfp-operator.git
cd kfp-operator

# Install from local chart
helm install kfp-operator ./helm/kfp-operator \
  --namespace kfp-operator-system \
  --create-namespace \
  --values values.yaml
```

## Verification

### Check Operator Status

```bash
# Check operator pods
kubectl get pods -n kfp-operator-system

# Expected output:
# NAME                                           READY   STATUS    RESTARTS   AGE
# kfp-operator-controller-manager-xxx-xxx        2/2     Running   0          2m

# Check Custom Resource Definitions
kubectl get crd | grep pipelines.kubeflow.org

# Expected output:
# pipelines.pipelines.kubeflow.org
# providers.pipelines.kubeflow.org
# runconfigurations.pipelines.kubeflow.org
# runs.pipelines.kubeflow.org
```

### Verify Operator Logs

```bash
# Check operator logs for any errors
kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager

# Should show successful startup messages
```

### Test Basic Functionality

```bash
# Create a test provider (adjust for your environment)
cat <<EOF | kubectl apply -f -
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Provider
metadata:
  name: test-provider
  namespace: default
spec:
  type: kfp
  kfp:
    restKfpApiUrl: "http://ml-pipeline.kubeflow.svc.cluster.local:8888"
EOF

# Check provider status
kubectl get providers
# Should show the provider with Ready status
```

## Configuration Options

### Core Configuration Parameters

| Parameter                | Description                   | Default | Production Recommendation |
|--------------------------|-------------------------------|---------|---------------------------|
| `manager.replicas`       | Number of controller replicas | 1       | 2+ for HA                 |
| `manager.resources`      | Resource requests/limits      | Basic   | Set based on cluster size |
| `logging.verbosity`      | Log level (0-2)               | 1       | 1 for production          |
| `statusFeedback.enabled` | Enable event system           | false   | true for automation       |

### Advanced Configuration

For detailed configuration options, see:
- [Configuration Reference](../configuration/) - Complete parameter documentation
- [Security Configuration](../security/) - RBAC and security settings
- [Monitoring Setup](../monitoring/) - Observability configuration

## Security Setup

### RBAC Configuration

The operator requires specific RBAC permissions. The Helm chart creates these automatically, but you can customize them:

```yaml
# Custom RBAC configuration
manager:
  rbac:
    create: true
    # Additional cluster roles for multi-tenancy
    additionalClusterRoles:
      - name: kfp-operator-tenant-viewer
        rules:
          - apiGroups: ["pipelines.kubeflow.org"]
            resources: ["pipelines", "runs"]
            verbs: ["get", "list", "watch"]
```

### Service Account Configuration

```yaml
# Service account configuration
manager:
  serviceAccount:
    create: true
    name: kfp-operator-controller-manager
    annotations:
      # For cloud provider integration
      iam.gke.io/gcp-service-account: kfp-operator@project.iam.gserviceaccount.com
```

## Monitoring Setup

### Prometheus Integration

If you have Prometheus Operator installed:

```yaml
# Enable ServiceMonitor for Prometheus
manager:
  monitoring:
    create: true
    serviceMonitor:
      create: true
      interval: 30s
      scrapeTimeout: 10s
```

### Grafana Dashboard

Import the KFP Operator Grafana dashboard:

```bash
# Download dashboard JSON
curl -o kfp-operator-dashboard.json \
  https://raw.githubusercontent.com/sky-uk/kfp-operator/master/monitoring/grafana-dashboard.json

# Import into Grafana UI or via ConfigMap
```

## Troubleshooting

### Common Installation Issues

#### 1. CRD Installation Failures
```bash
# Check CRD status
kubectl get crd | grep pipelines.kubeflow.org

# If missing, manually install
kubectl apply -f https://raw.githubusercontent.com/sky-uk/kfp-operator/master/config/crd/bases/
```

#### 2. Webhook Certificate Issues
```bash
# Check webhook configuration
kubectl get validatingwebhookconfiguration | grep kfp-operator

# Check certificate status (if using cert-manager)
kubectl get certificates -n kfp-operator-system
```

#### 3. RBAC Permission Errors
```bash
# Check service account permissions
kubectl auth can-i create pipelines --as=system:serviceaccount:kfp-operator-system:kfp-operator-controller-manager

# Review RBAC configuration
kubectl get clusterroles | grep kfp-operator
```

### Getting Help

- [Troubleshooting Guide](../troubleshooting/) - Detailed solutions
- Review operator logs for specific error messages
- Verify prerequisites are properly installed
- [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues) - Known problems
- [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions) - Community help

## Next Steps

After successful installation:

- [Configure Providers](../configuration/providers/) - Set up ML platform connections
- [Set up Security](../security/) - Implement RBAC and security policies
- [Configure Monitoring](../monitoring/) - Set up comprehensive observability
- [Test with ML Engineers](../../ml-engineers/getting-started/) - Verify user workflows

---

**Next:** [Provider Configuration](../configuration/providers/).
