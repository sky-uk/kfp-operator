---
title: "Documentation (master)"
linkTitle: "Documentation"
description: "Comprehensive documentation for the Kubeflow Pipelines Operator - bringing GitOps and declarative management to your ML workflows"
weight: 10
---

# KFP Operator Documentation

The Kubeflow Pipelines Operator provides a **Kubernetes-native API** for Kubeflow and VertexAI pipelines. Define and manage ML pipelines as code using kubectl.

## Key Features
- 
- **Infrastructure as Code**: Apply Kubernetes patterns to ML workflows
- **Event-Driven**: Automated pipeline execution and management
- **Enterprise Ready**: RBAC, security policies, and multi-tenant isolation
- **Developer Friendly**: Use kubectl, Helm, and existing CI/CD pipelines

## Choose Your Path

### [For ML Engineers & Data Scientists](ml-engineers/)
**Build and deploy ML pipelines using the KFP Operator**

Choose this path if you:
- Develop and deploy machine learning pipelines
- Use TFX, Kubeflow Pipelines, or similar ML frameworks
- Need to run experiments and manage model training
- Want to automate ML workflows with GitOps

**Includes:**
- Quick Start Guides
- Practical Tutorials
- Best Practices
- API Reference
- Troubleshooting

### [For Platform Engineers & Developers](platform-engineers/)
**Install, configure, and maintain the KFP Operator platform**

Choose this path if you:
- Install and configure the KFP Operator in Kubernetes clusters
- Manage platform infrastructure and operations
- Develop extensions or contribute to the operator
- Set up multi-tenant ML platforms for teams

**Includes:**
- Installation Guides
- Architecture Deep-Dives
- Configuration Reference
- Security & RBAC
- Maintenance & Operations

### [API Reference](reference/)
**Complete technical reference for all Custom Resource Definitions**

For developers and advanced users who need:
- Complete API specifications and CRDs
- Technical documentation cross-references
- Consolidated access to all specifications

## Architecture

The KFP Operator extends Kubernetes with custom resources for ML pipeline entities:

- **Custom Resources**: Kubernetes-native representations of pipelines, runs, and configurations
- **Controller**: Manages resource lifecycle and orchestrates workflows
- **Provider Service**: Abstracts different ML platforms (KFP, Vertex AI)
- **Event System**: Publishes pipeline events for reactive workflows

## Community and Support

- **Discussions**: [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions) for questions and ideas
- **Issues**: [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues) for bug reports and feature requests
- **Releases**: [GitHub Releases](https://github.com/sky-uk/kfp-operator/releases) for latest versions

## Contributing

Open source project welcoming contributions from ML practitioners and platform engineers.

- **Source Code**: [GitHub Repository](https://github.com/sky-uk/kfp-operator)
- **Contributing Guide**: See `CONTRIBUTING.md` in the repository
