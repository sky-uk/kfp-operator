---
title: "For ML Engineers & Data Scientists"
linkTitle: "ML Engineers"
description: "Documentation for teams who build and deploy ML pipelines using the KFP Operator"
weight: 10
---

# Documentation for ML Engineers & Data Scientists

Documentation for teams who build, deploy, and manage machine learning pipelines using the KFP Operator.

## Who This Is For

This documentation is for you if you:

- Develop and train machine learning models
- Run experiments with different algorithms, hyperparameters, and datasets
- Deploy and automate ML workflows from data ingestion to model serving
- Manage model versioning, validation, and deployment

## Prerequisites

### Required Knowledge
- Basic machine learning concepts
- Python programming
- Basic Kubernetes (pods, services, kubectl)

### Technical Requirements
- Kubernetes cluster with KFP Operator installed
- kubectl configured to access your cluster
- Container registry access (Docker Hub, GCR, etc.)
- ML framework (TFX, Kubeflow Pipelines, or similar)

### Optional
- Docker for building custom pipeline images
- Git for version controlling pipeline code

## Quick Start Checklist

- [ ] Verify access to your Kubernetes cluster
- [ ] Confirm KFP Operator is installed and running
- [ ] Set up your development environment
- [ ] Choose a tutorial based on your use case
- [ ] Deploy your first pipeline
- [ ] Explore additional examples and patterns

## Common Use Cases

### Model Training & Experimentation
- A/B testing different model architectures
- Continuous model retraining on new data
- Distributed training across multiple nodes

### Model Deployment & Serving
- Integration with serving platforms

### MLOps & Automation
- Event-driven pipeline execution
- Integration with CI/CD systems
- Multi-environment promotion workflows

## Getting Help

- [Troubleshooting](troubleshooting/) - Common problems and solutions
- [Best Practices](best-practices/) - Recommended patterns
- [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues) - Known issues
- [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions) - Community help
- [Platform Engineer docs](../platform-engineers/) - Installation/configuration issues

## Related Resources

- [TFX Documentation](https://www.tensorflow.org/tfx) - TensorFlow Extended framework
- [Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/) - ML workflow platform
- [Kubernetes Documentation](https://kubernetes.io/docs/) - Container orchestration

---

**Next:** [Getting Started](getting-started/) guide.
