---
title: "Configuration"
linkTitle: "Configuration"
description: "Advanced configuration options for the KFP Operator platform"
weight: 20
---

# Platform Configuration

This section provides comprehensive guidance for configuring the KFP Operator platform for production use. As a platform engineer, you'll learn how to customize the operator behavior, set up providers, and optimize for your specific environment.

## 📋 Configuration Overview

The KFP Operator configuration consists of several layers:

### 1. **Operator Configuration** 
Core operator settings and behavior
- [Operator Configuration](operator-configuration/) - Main operator settings
- Resource allocation and scaling
- Logging and monitoring configuration
- Webhook and admission control settings

### 2. **Provider Configuration**
ML orchestration platform integrations
- [Providers Overview](providers/) - Provider system architecture
- [Kubeflow Pipelines Provider](providers/kfp/) - KFP integration
- [Vertex AI Provider](providers/vertex/) - Google Cloud integration
- [Custom Providers](providers/custom/) - Building new providers

### 3. **Security Configuration**
RBAC, policies, and compliance settings
- [Security Setup](../security/) - Enterprise security patterns
- Service account configuration
- Network policies and isolation
- Secret and credential management

### 4. **Monitoring Configuration**
Observability and alerting setup
- [Monitoring Setup](../monitoring/) - Comprehensive observability
- Metrics collection and export
- Log aggregation and analysis
- Alerting rules and notifications

## Quick Configuration Guide

### Basic Production Setup

#### 1. Operator Configuration
```yaml
# values.yaml - Production configuration
manager:
  replicas: 2  # High availability
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  
  # Enable monitoring
  monitoring:
    create: true
    serviceMonitor:
      create: true

# Enable event system
statusFeedback:
  enabled: true

# Production logging
logging:
  verbosity: 1
```

#### 2. Provider Setup
```yaml
# kubeflow-provider.yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: production-kfp
  namespace: kfp-operator-system
spec:
  type: kfp
  kfp:
    restKfpApiUrl: "https://kubeflow.company.com/pipeline"
    uiUrl: "https://kubeflow.company.com/"
```

#### 3. RBAC Configuration
```yaml
# Enable RBAC with custom roles
manager:
  rbac:
    create: true
    additionalClusterRoles:
      - name: kfp-operator-viewer
        rules:
          - apiGroups: ["pipelines.kubeflow.org"]
            resources: ["pipelines", "runs"]
            verbs: ["get", "list", "watch"]
```

### Environment-Specific Configurations

#### Development Environment
```yaml
# dev-values.yaml
manager:
  replicas: 1
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

logging:
  verbosity: 2  # Debug logging

# Relaxed timeouts for development
argo:
  stepTimeoutSeconds:
    default: 3600  # 1 hour
```

#### Staging Environment
```yaml
# staging-values.yaml
manager:
  replicas: 1
  resources:
    requests:
      cpu: 150m
      memory: 192Mi
    limits:
      cpu: 750m
      memory: 768Mi

# Production-like settings with shorter retention
argo:
  ttlStrategy:
    secondsAfterCompletion: 1800  # 30 minutes
```

#### Production Environment
```yaml
# prod-values.yaml
manager:
  replicas: 3  # High availability
  resources:
    requests:
      cpu: 300m
      memory: 512Mi
    limits:
      cpu: 1500m
      memory: 2Gi

# Production monitoring
monitoring:
  create: true
  serviceMonitor:
    create: true
    interval: 30s

# Strict security
rbac:
  create: true
webhookCertificates:
  provider: cert-manager
```

## 🆘 Configuration Troubleshooting

### Common Configuration Issues

#### Invalid Helm Values
```bash
# Debug Helm template rendering
helm template kfp-operator kfp-operator/kfp-operator \
  --values values.yaml \
  --debug \
  --output-dir ./debug-output

# Check for YAML syntax errors
yamllint values.yaml
```

#### Provider Connection Issues
```bash
# Test provider connectivity
kubectl run debug-pod --image=curlimages/curl --rm -it -- \
  curl -v http://kubeflow-provider-url/api/v1/healthz

# Check DNS resolution
kubectl run debug-pod --image=busybox --rm -it -- \
  nslookup kubeflow-provider-hostname
```

### Getting Help

For configuration issues:
1. **Check [Troubleshooting Guide](../troubleshooting/)** for platform issues
2. **Review [Architecture Documentation](../architecture/)** for system understanding
3. **Consult [Security Guide](../security/)** for RBAC and security issues
4. **Ask in [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)** for community help

---

**Ready to configure your platform?** Start with the [Operator Configuration](operator-configuration/) guide for core settings, then move on to [Provider Setup](providers/) for ML platform integration.
