# Kubeflow Pipelines Operator

The Kubeflow Pipelines Operator (KFP Operator) provides a declarative API for managing and running machine learning pipelines across multiple providers using Kubernetes [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). It streamlines the deployment, execution, and lifecycle management of ML pipelines while promoting best engineering practices and reducing operational overhead.

## Features

- **Declarative Pipeline Management**: Define ML pipelines as Kubernetes resources with full lifecycle management
- **Multi-Provider Support**: Run pipelines on various orchestration providers including Vertex AI and KFP standalone
- **Framework Support**: Compatible with [TFX](https://www.tensorflow.org/tfx) and [KFP SDK](https://kubeflow-pipelines.readthedocs.io/) pipelines
- **Event-Driven Architecture**: Optional integration with [Argo Events](https://argoproj.github.io/argo-events/) for reactive pipeline execution
- **Kubernetes-Native**: Follows standard Kubernetes operator patterns with custom controllers
- **Version Management**: Automatic pipeline versioning and artifact management
- **Environment Isolation**: Provider-specific configuration without manual setup

## Architecture Overview

The KFP Operator follows the standard Kubernetes operator pattern where controllers manage the state of custom resources. Each controller creates [Argo Workflows](https://argoproj.github.io/workflows/) that interact with Provider Services, which in turn communicate with orchestration providers (e.g., Vertex AI).

**Key Components:**
- **Controllers**: Manage Pipeline, Run, Provider, and Experiment resources
- **Provider Service**: Abstracts communication with different ML orchestration platforms
- **Argo Workflows**: Execute pipeline operations and provider interactions
- **Custom Resources**: Define pipelines, runs, providers, and experiments declaratively

## Custom Resources

The operator provides several custom resources:

- **Pipeline**: Defines ML pipeline lifecycle and configuration
- **Run**: Represents individual pipeline executions
- **Provider**: Configures orchestration provider settings
- **Experiment**: Groups related pipeline runs
- **RunSchedule**: Schedules recurring pipeline executions
- **RunConfiguration**: Templates for pipeline runs

## Supported Providers

- **Vertex AI**: Google Cloud's managed ML platform
- **KFP V2 Standalone**: Self-hosted Kubeflow Pipelines

## Requirements

- **Kubernetes**: v1.19 or later
- **Python**: 3.7, 3.9 (for TFX pipelines)
- **Dependencies**: Argo Workflows

## Documentation

- **[Complete Documentation](https://sky-uk.github.io/kfp-operator)**: Comprehensive guides and API reference
- **[Getting Started](https://sky-uk.github.io/kfp-operator/docs/getting-started/overview/)**: Detailed installation and setup instructions
- **[API Reference](https://sky-uk.github.io/kfp-operator/docs/reference/resources/)**: Custom resource specifications
- **[Provider Configuration](https://sky-uk.github.io/kfp-operator/docs/reference/providers/overview/)**: Provider setup guides

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- Reporting bugs and requesting features
- Development setup and guidelines
- Submitting pull requests
- Code style and testing requirements

For development setup, see the [Development Guide](DEVELOPMENT.md).

## Support

- **Issues**: Report bugs and request features via [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues)
- **Discussions**: Join conversations in [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)
- **Documentation**: Visit our [documentation site](https://sky-uk.github.io/kfp-operator) for detailed guides
