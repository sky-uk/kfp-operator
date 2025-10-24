---
title: "For Platform Engineers & Developers"
linkTitle: "Platform Engineers"
description: "Documentation for teams who install, configure, and maintain the KFP Operator platform"
weight: 20
---

# Documentation for Platform Engineers & Developers

Documentation for teams who install, configure, maintain, and extend the KFP Operator platform.

## Who This Is For

This documentation is for you if you:

- Deploy and configure the KFP Operator in Kubernetes clusters
- Manage platform services and infrastructure
- Implement security, RBAC, and compliance policies
- Set up monitoring, logging, and alerting
- Develop extensions or contribute to the operator
- Set up multi-tenant platforms for ML teams

## What You'll Learn

- Installation & deployment in various environments
- Advanced operator configuration and tuning
- System architecture and internals
- Enterprise security patterns and RBAC
- Production monitoring and maintenance
- Development and platform extension

## Implementation Path

### 1. Installation & Setup
Get the operator running in your cluster:

- [Installation Guide](installation/) - Deploy the operator with Helm
- [Configuration](configuration/) - Configure for your environment
- [Provider Setup](providers/) - Connect to ML orchestration platforms

### 2. Security & RBAC
Implement enterprise security and access control:

- [Security Configuration](security/) - Secure the operator deployment
- [RBAC Setup](rbac/) - Configure role-based access control
- [Multi-tenancy](multi-tenancy/) - Isolate teams and projects

### 3. Operations & Monitoring
Set up production monitoring and maintenance:

- [Monitoring Setup](monitoring/) - Metrics, logging, and alerting
- [Maintenance](maintenance/) - Upgrades, backups, and operations
- [Troubleshooting](troubleshooting/) - Debug operator issues

### 4. Advanced Topics
Master advanced features and customization:

- [Architecture Deep Dive](architecture/) - Understand system internals
- [Custom Providers](development/providers/) - Build new provider integrations
- [Contributing](development/contributing/) - Contribute to the project

## Prerequisites

### Required Knowledge
- Kubernetes cluster administration
- Helm charts and package management
- Docker and container registries
- YAML and configuration management

### Technical Requirements
- Kubernetes cluster admin access (v1.21+)
- Helm 3.x for installing and managing the operator
- kubectl configured with cluster admin permissions
- Container registry for storing operator and pipeline images

### Recommended Experience
- Kubernetes operators and controllers
- GitOps workflows and tools
- Prometheus, Grafana, and observability
- Continuous integration and deployment

## Quick Deployment Checklist

- [ ] Verify Kubernetes cluster meets requirements
- [ ] Install dependencies (Argo Workflows and Argo Events)
- [ ] Add KFP Operator Helm repository
- [ ] Create values.yaml for your environment
- [ ] Install using Helm with custom configuration
- [ ] Verify all components are running correctly
- [ ] Configure providers and connections to ML platforms
- [ ] Deploy a test pipeline to verify functionality

## Deployment Scenarios

- **Development**: Single-node clusters (minikube, kind), minimal resources
- **Production**: Multi-node HA clusters, advanced security, monitoring
- **Cloud**: Managed K8s services (GKE, EKS, AKS), auto-scaling
- **GitOps**: ArgoCD/Flux deployment, Infrastructure as Code

## Architecture Overview

Key components:

- **Controller Manager**: Core operator logic and reconciliation
- **Admission Webhooks**: Validation and mutation of resources
- **Provider Services**: Abstraction layer for ML platforms
- **Argo Integration**: Workflow execution and management
- **Event System**: Pipeline event processing and distribution

## Security Considerations

### Access Control
- RBAC for operator components
- Service account management
- Network policies and segmentation
- Secret and credential management

### Data Security
- Encryption at rest and in transit
- Secure artifact storage
- Pipeline data isolation
- Audit logging and compliance

### Threat Mitigation
- Container image scanning
- Runtime security monitoring
- Vulnerability management
- Incident response procedures

## Getting Help

- [Troubleshooting](troubleshooting/) - Common platform issues
- [Architecture](architecture/) - System behavior and internals
- [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues) - Known problems
- [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions) - Community help
- [ML Engineer docs](../ml-engineers/) - User-facing issues

## Related Resources

### Kubernetes Ecosystem
- [Argo Workflows](https://argoproj.github.io/argo-workflows/) - Workflow execution engine
- [Argo Events](https://argoproj.github.io/argo-events/) - Event-driven automation
- [Prometheus Operator](https://prometheus-operator.dev/) - Monitoring and alerting

### Platform Engineering
- [CNCF Landscape](https://landscape.cncf.io/) - Cloud native technologies
- [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) - Operator pattern
- [GitOps](https://www.gitops.tech/) - GitOps principles and tools

---

**Next:** [Installation Guide](installation/).
