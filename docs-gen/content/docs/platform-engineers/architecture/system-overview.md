---
title: "System Overview"
linkTitle: "System Overview"
description: "Comprehensive overview of KFP Operator architecture and system design for platform engineers"
weight: 10
---

# Introduction to the KFP Operator

The **Kubeflow Pipelines Operator** is a Kubernetes-native solution that brings GitOps and declarative management to machine learning workflows. Instead of manually managing pipelines through UIs or imperative scripts, you can define your entire ML infrastructure as code using familiar Kubernetes patterns.

## Why Use the KFP Operator?

### Traditional ML Pipeline Management Challenges

- **Manual Deployment**: Uploading pipelines through web UIs or scripts
- **Configuration Drift**: Environment-specific settings scattered across systems
- **Limited Automation**: Manual intervention required for pipeline lifecycle
- **Inconsistent Environments**: Different configurations between dev/staging/prod

### KFP Operator Solutions

- **Infrastructure as Code**: Declarative pipeline definitions using Kubernetes CRDs
- **Automated Lifecycle**: Event-driven pipeline execution and management
- **Enterprise Security**: RBAC, policies, and multi-tenant isolation
- **Developer Experience**: Use familiar tools like `kubectl`, Helm, and CI/CD

## Core Concepts

### Custom Resource Definitions (CRDs)

The operator extends Kubernetes with custom resources that represent ML pipeline entities:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Pipeline
metadata:
  name: training-pipeline
  namespace: ml-team
spec:
  image: my-org/ml-pipeline:v1.2.0
  env:
    - name: MODEL_VERSION
      value: "v2.1"
```

### Declarative Management

Define your desired state, and the operator ensures your cluster matches:

```bash
# Deploy pipeline
kubectl apply -f pipeline.yaml

# Check status
kubectl get pipelines

# Update configuration
kubectl patch pipeline training-pipeline --patch '{"spec":{"env":[{"name":"MODEL_VERSION","value":"v2.2"}]}}'
```

## Compatibility

The operator currently supports:

### ML Frameworks
- **TFX Pipelines** with Python 3.7, 3.8, and 3.9
- **[Kubeflow Pipelines SDK](https://kubeflow-pipelines.readthedocs.io/)** v1.8+
- **Custom ML frameworks** via container-based components

### Orchestration Platforms
- **Kubeflow Pipelines** (self-hosted or managed)
- **Vertex AI** (Google Cloud's managed ML platform)
- **Extensible provider system** for additional platforms

### Kubernetes Versions
- **Kubernetes** v1.21+ (tested up to v1.28)

### Key Benefits

- **Focus on Logic**: Write pipeline logic, not infrastructure code
- **Automatic Configuration**: Environment setup handled by the operator
- **Portable Pipelines**: Same code works across environments
- **Faster Development**: Reduced boilerplate and setup time

### Example Pipeline

Here's a complete example of a TFX pipeline optimized for the KFP Operator:

```python
# penguin_pipeline.py
import tensorflow_model_analysis as tfma
from tfx import v1 as tfx

def create_pipeline(
    pipeline_name: str,
    pipeline_root: str,
    data_root: str,
    module_file: str,
    serving_model_dir: str,
    metadata_path: str,
) -> tfx.dsl.Pipeline:
    """Creates a TFX pipeline for penguin classification."""

    # Data ingestion
    example_gen = tfx.components.CsvExampleGen(
        input_base=data_root
    )

    # Data validation
    statistics_gen = tfx.components.StatisticsGen(
        examples=example_gen.outputs['examples']
    )

    schema_gen = tfx.components.SchemaGen(
        statistics=statistics_gen.outputs['statistics'],
        infer_feature_shape=True
    )

    example_validator = tfx.components.ExampleValidator(
        statistics=statistics_gen.outputs['statistics'],
        schema=schema_gen.outputs['schema']
    )

    # Feature engineering
    transform = tfx.components.Transform(
        examples=example_gen.outputs['examples'],
        schema=schema_gen.outputs['schema'],
        module_file=module_file
    )

    # Model training
    trainer = tfx.components.Trainer(
        module_file=module_file,
        examples=transform.outputs['transformed_examples'],
        transform_graph=transform.outputs['transform_graph'],
        schema=schema_gen.outputs['schema'],
        train_args=tfx.proto.TrainArgs(num_steps=2000),
        eval_args=tfx.proto.EvalArgs(num_steps=5)
    )

    # Model evaluation
    model_resolver = tfx.components.Resolver(
        strategy_class=tfx.dsl.experimental.LatestBlessedModelStrategy,
        model=tfx.dsl.Channel(type=tfx.types.standard_artifacts.Model),
        model_blessing=tfx.dsl.Channel(
            type=tfx.types.standard_artifacts.ModelBlessing
        )
    ).with_id('latest_blessed_model_resolver')

    eval_config = tfma.EvalConfig(
        model_specs=[tfma.ModelSpec(label_key='species')],
        slicing_specs=[tfma.SlicingSpec()],
        metrics_specs=[
            tfma.MetricsSpec(metrics=[
                tfma.MetricConfig(class_name='SparseCategoricalAccuracy'),
                tfma.MetricConfig(class_name='ExampleCount')
            ])
        ]
    )

    evaluator = tfx.components.Evaluator(
        examples=example_gen.outputs['examples'],
        model=trainer.outputs['model'],
        baseline_model=model_resolver.outputs['model'],
        eval_config=eval_config
    )

    # Model deployment
    pusher = tfx.components.Pusher(
        model=trainer.outputs['model'],
        model_blessing=evaluator.outputs['blessing'],
        push_destination=tfx.proto.PushDestination(
            filesystem=tfx.proto.PushDestination.Filesystem(
                base_directory=serving_model_dir
            )
        )
    )

    return tfx.dsl.Pipeline(
        pipeline_name=pipeline_name,
        pipeline_root=pipeline_root,
        components=[
            example_gen,
            statistics_gen,
            schema_gen,
            example_validator,
            transform,
            trainer,
            model_resolver,
            evaluator,
            pusher,
        ],
        enable_cache=True,
        metadata_connection_config=tfx.orchestration.metadata.sqlite_metadata_connection_config(
            metadata_path
        )
    )
```

For a complete working example, see the [penguin pipeline]({{< param "github_repo" >}}/blob/{{< param "github_branch" >}}/{{< param "github_subdir" >}}/includes/master/quickstart/penguin_pipeline/pipeline.py) in our repository.

## Parameters

### Parameter Types

The KFP Operator supports multiple parameter types for maximum flexibility:

| Parameter Type              | Lifecycle Phase | Description                                      | Use Cases                                      | Example                       |
|-----------------------------|-----------------|--------------------------------------------------|------------------------------------------------|-------------------------------|
| **Named Constants**         | Development     | Hard-coded values in pipeline definition         | Model architecture, fixed configurations       | `HIDDEN_UNITS = 128`          |
| **Compile-time Parameters** | Compilation     | Environment variables applied during compilation | Data sources, model paths, environment configs | `DATA_ROOT`, `MODEL_REGISTRY` |
| **Runtime Parameters**      | Execution       | Values that can change between runs              | Hyperparameters, experiment settings           | `learning_rate`, `num_epochs` |

### Parameter Implementation Examples

#### Compile-time Parameters
```yaml
# In your Pipeline resource
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Pipeline
metadata:
  name: training-pipeline
spec:
  image: my-org/pipeline:v1.0.0
  env:
    - name: DATA_ROOT
      value: "gs://production-bucket/data"
    - name: MODEL_REGISTRY
      value: "gs://model-registry/models"
```

#### Runtime Parameters
```python
# In your pipeline definition
from tfx.v1.dsl.experimental import RuntimeParameter

learning_rate = RuntimeParameter(
    name='learning_rate',
    default=0.001,
    ptype=float
)

num_epochs = RuntimeParameter(
    name='num_epochs',
    default=10,
    ptype=int
)
```

```yaml
# In your Run resource
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Run
metadata:
  name: experiment-run-1
spec:
  pipeline: training-pipeline
  runtimeParameters:
    learning_rate: "0.01"
    num_epochs: "20"
```

### Best Practices

- **Use Compile-time Parameters** for environment-specific settings
- **Use Runtime Parameters** for experimentation and hyperparameter tuning
- **Document Parameters** clearly in your pipeline code and CRD specs
- **Version Parameters** alongside your pipeline code for reproducibility

## Event-Driven Workflows

The KFP Operator provides powerful event-driven capabilities through integration with [Argo Events](https://argoproj.github.io/argo-events/), enabling reactive ML workflows that respond to triggers.

### Why Event-Driven ML?

Traditional ML workflows often require manual intervention when a model finishes training. Event-driven workflows enable:
- **Continuous Deployment**: Deploy models automatically after successful training to serving
- **Reactive Pipelines**: Trigger pipelines based on events (dependent pipeline completions, schedules, pipeline / runconfiguration changes)
- And much more!

### Supported Event Sources

#### Run Completion Events
React to pipeline execution completion with detailed status information:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: RunConfiguration
metadata:
  name: retrain-on-dependency-completion
spec:
  run:
    pipeline: training-pipeline
  triggers:
    onChange:
    - pipeline
    - runSpec
    runConfigurations:
    - some-namespace/some-dependency-rc
```

**Use Cases:**
- Trigger downstream pipelines after successful completion
- Send notifications on pipeline failures
- Archive artifacts after pipeline completion
- Deliver notification to serving platform to use the latest model

For detailed information, see the [Run Completion Events Reference](../events/run-completion).