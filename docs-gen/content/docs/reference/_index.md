---
title: "API Reference"
linkTitle: "Reference"
description: "Comprehensive API reference and technical specifications for the KFP Operator"
weight: 30
---

# KFP Operator API Reference

This section provides the canonical, comprehensive API reference documentation for the Kubeflow Pipelines Operator. It serves as the authoritative technical documentation for all Custom Resource Definitions (CRDs), configuration options, and system interfaces.

## Intended Audience

This reference is designed for:

- **API Developers**: Building integrations with the KFP Operator
- **Advanced Users**: Requiring detailed technical specifications
- **Contributors**: Developing features or fixing bugs
- **Platform Architects**: Designing systems that interact with the operator
- **Technical Writers**: Creating user-focused documentation

## Reference Sections

### [Custom Resources](resources/)
Complete API specifications for all Kubernetes Custom Resource Definitions (CRDs) provided by the KFP Operator.

**Available Resources:**
- **[Pipeline](resources/pipeline/)** - Reusable ML pipeline templates
- **[Run](resources/run/)** - Individual pipeline executions
- **[RunConfiguration](resources/runconfiguration/)** - Automated pipeline execution
- **[Provider](resources/provider/)** - ML orchestration platform connections

## Additional Technical Documentation

The following technical documentation has been organized by user type for better accessibility:

### For Platform Engineers
- **[Configuration Reference](../platform-engineers/configuration/)** - Operator configuration and provider settings
- **[Architecture Documentation](../platform-engineers/architecture/)** - System design and technical specifications
- **[Event System](../platform-engineers/events/)** - Event schemas and integration patterns

### For ML Engineers
- **[Framework Integrations](../ml-engineers/frameworks/)** - ML framework compatibility and specifications

## API Versioning

- **Current Version**: `pipelines.kubeflow.org/v1beta1`
- **Stability**: Beta (API stable, production ready)
- **Compatibility**: See individual resource documentation for version history and compatibility details

## Related Documentation

### User-Focused Guides
- **[ML Engineers Documentation](../ml-engineers/)**: Pipeline development and usage guides
- **[Platform Engineers Documentation](../platform-engineers/)**: Installation and operations guides

### External References
- **[Kubernetes API Conventions](https://kubernetes.io/docs/reference/using-api/api-concepts/)**: Kubernetes API patterns
- **[Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)**: CRD concepts
- **[Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)**: Kubernetes operator pattern

---

**Navigate to specific sections above for complete technical specifications, examples, and implementation details.**
