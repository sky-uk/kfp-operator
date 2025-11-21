---
title: "Getting Started"
linkTitle: "Getting Started"
description: "Quick start guide for ML engineers to deploy their first pipeline with the KFP Operator"
weight: 1
---

# Getting Started with ML Pipelines

This guide will help you deploy your first ML pipeline using the KFP Operator.

## What You'll Do

- Deploy a pipeline using Kubernetes resources
- Run the pipeline with custom parameters
- Set up automated scheduling
- Monitor pipeline status

## Prerequisites

- Access to a Kubernetes cluster with KFP Operator installed
- kubectl configured to access your cluster
- Basic familiarity with Kubernetes concepts (pods, services)
- Container registry access (Docker Hub, GCR, etc.) for custom images

> **Need the operator installed?** Check with your platform team or see the [Platform Engineers documentation](../../platform-engineers/installation/) for installation instructions.

## Core Concepts

The KFP Operator extends Kubernetes with custom resources for ML pipelines:

- **Pipeline**: Defines a reusable ML pipeline template
- **Run**: Represents a single execution of a pipeline
- **RunConfiguration**: Configures automated pipeline execution
- **Provider**: Manages connections to ML orchestration platforms

Instead of uploading pipelines through UIs, you define them as configuration files:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: my-training-pipeline
spec:
  provider: provider-namespace/provider-name
  image: "my-registry/ml-pipeline:v1.0.0"
  framework:
    name: tfx
    parameters:
      pipeline: my_pipeline.create_components
```

## Quick Start Tutorial

### Step 1: Verify Your Environment

First, check that the KFP Operator is running in your cluster:

```bash
# Check if the operator is installed
kubectl get pods -n kfp-operator-system

# Verify Custom Resource Definitions are available
kubectl get crd | grep pipelines.kubeflow.org
```

You should see the operator pods running and several CRDs listed.

### Step 2: Check Available Providers

Providers connect the operator to ML orchestration platforms:

```bash
# List available providers
kubectl get providers

# Example output:
# NAME                STATUS   TYPE   AGE
# kubeflow-provider   Ready    kfp    5m
```

If no providers are available, contact your platform team to set one up.

### Step 3: Create Your First Pipeline

Create a file called `my-first-pipeline.yaml`:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: penguin-training
  namespace: default
spec:
  provider: provider-namespace/provider-name
  image: "gcr.io/kfp-operator/penguin-pipeline:latest"
  framework:
    name: tfx
    parameters:
      pipeline: penguin_pipeline.create_components
  env:
    - name: DATA_ROOT
      value: "gs://kfp-operator-examples/penguin-data"
    - name: MODEL_ROOT
      value: "gs://my-bucket/models"  # Replace with your bucket
```

Deploy the pipeline:

```bash
kubectl apply -f my-first-pipeline.yaml
```

Check the pipeline status:

```bash
kubectl get pipelines
# NAME              STATUS   PROVIDER            AGE
# penguin-training  Ready    kubeflow-provider   30s
```

### Step 5: Set Up Automated Execution

Create automated pipeline execution with `run-configuration.yaml`:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: RunConfiguration
metadata:
  name: daily-training
  namespace: default
spec:
  run:
    provider: provider-namespace/provider-name
    pipeline: penguin-training
    parameters:
      - name: num_epochs
        value: "10"
      - name: learning_rate
        value: "0.001"
  triggers:
    schedules:
      - cronExpression: "0 2 * * *"  # Daily at 2 AM
        startTime: "2024-01-01T00:00:00Z"
        endTime: "2024-12-31T23:59:59Z"
    onChange:
      - pipeline
```

Apply the configuration:

```bash
kubectl apply -f run-configuration.yaml

# Verify it's scheduled
kubectl get runconfigurations
```

## Complete!

You've successfully:

- Deployed a pipeline using Kubernetes resources
- Set up automated daily training
- Executed a pipeline run with custom parameters
- Learned how to monitor pipeline execution

## Understanding What Happened

When you created the Pipeline resource, the KFP Operator:

1. Validated your pipeline specification
2. Compiled the pipeline for your target platform
3. Registered it with your ML orchestration platform (Kubeflow Pipelines)
4. Updated the status to show it's ready

When you created the Run Configuration resource, the operator:

1. Created a RunSchedule to schedule the workflow for future runs
2. Created a Run resource to execute the pipeline
3. Monitored execution and updated status
4. Published events for downstream automation

### Key Benefits

- **Version Control**: Your pipeline definitions are now in YAML files and can be versioned in Git
- **Reproducible**: Same pipeline definition works across environments
- **Automated**: Set-and-forget scheduling with RunConfiguration
- **Observable**: Full visibility through kubectl and Kubernetes events

## What's Next?

- [Build Custom Pipelines](./tutorials/training-pipeline): Create your own TFX pipelines
- [Pipeline Dependencies](./tutorials/pipeline-dependencies): Chain multiple pipelines together
- [Best Practices](../best-practices/): Production ML engineering patterns
- [API Reference](../../reference/): Complete Custom Resource specifications
- [Troubleshooting](../troubleshooting/): Debug common pipeline issues

## Need Help?

If you encounter issues:

1. Check [Troubleshooting](../troubleshooting/) for common problems
2. Review pipeline logs: `kubectl logs -l workflows.argoproj.io/workflow=<workflow-name>`
3. Check operator logs: `kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager`
4. Ask for help: [GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)

---

**Next:** [Training Pipeline Tutorial](../tutorials/training-pipeline/).
