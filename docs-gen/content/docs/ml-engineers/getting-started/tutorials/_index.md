---
title: "Tutorials"
linkTitle: "Tutorials"
description: "Step-by-step tutorials for building and deploying ML pipelines with the KFP Operator"
weight: 20
---

# ML Pipeline Tutorials

This section provides comprehensive, hands-on tutorials for building and deploying machine learning pipelines using the KFP Operator. Each tutorial includes complete code examples, step-by-step instructions, and best practices.

## 🗺️ Recommended Learning Path
1. **[Training Pipeline](training-pipeline/)** - Build a complete workflow
2. **[Pipeline Dependencies](pipeline-dependencies/)** - Learn chaining

## 🛠️ Tutorial Prerequisites

### Required for All Tutorials
- **Kubernetes cluster** with KFP Operator installed
- **kubectl** configured to access your cluster
- **Basic ML knowledge** and Python programming skills
- **Container registry access** for storing pipeline images

### Tutorial-Specific Requirements
- **Docker**: Required for tutorials involving custom pipeline building
- **Git**: Recommended for version controlling your pipeline code
- **Cloud Storage**: Some tutorials require cloud storage (GCS, S3, etc.)
- **ML Frameworks**: TFX knowledge helpful for advanced tutorials

## 📁 Code Repository

All tutorial code is available in our GitHub repository:

**[Tutorial Examples](https://github.com/sky-uk/kfp-operator/tree/master/includes/master)**

```bash
# Clone the repository to follow along
git clone https://github.com/sky-uk/kfp-operator.git
cd kfp-operator/docs-gen/includes/master
```

## 🆘 Getting Help

If you encounter issues while following the tutorials:

1. **Check the troubleshooting section** in each tutorial
2. **Review [Common Issues](../troubleshooting/)** for known problems
3. **Search [GitHub Issues](https://github.com/sky-uk/kfp-operator/issues)** for similar problems
4. **Ask in [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)** for community help
5. **Check [Platform Engineer docs](../../platform-engineers/)** for installation issues

## 🔗 Related Resources

### External Tutorials
- **[TFX Tutorials](https://www.tensorflow.org/tfx/tutorials)** - TensorFlow Extended framework
- **[Kubeflow Pipelines Tutorials](https://www.kubeflow.org/docs/components/pipelines/tutorials/)** - ML workflow platform
- **[Kubernetes Tutorials](https://kubernetes.io/docs/tutorials/)** - Container orchestration

---

**Ready to start building?** Choose a tutorial that matches your experience level and dive in!
