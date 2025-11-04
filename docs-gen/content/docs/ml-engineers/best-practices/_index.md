---
title: "Best Practices"
linkTitle: "Best Practices"
description: "Production-ready patterns and best practices for ML pipeline development"
weight: 50
---

# ML Pipeline Best Practices

Proven best practices for developing, deploying, and maintaining production ML pipelines using the KFP Operator.

## Pipeline Development

### Design Principles

#### 1. Modular and Reusable Design
**Create pipelines that can be easily modified and reused:**

```python
# ✅ Good: Modular pipeline with configurable components
def create_pipeline(
    data_root: str,
    model_root: str,
    preprocessing_config: str = "default",
    training_config: str = "default"
) -> tfx.dsl.Pipeline:
    # Configurable data ingestion
    example_gen = create_example_gen(data_root, preprocessing_config)
    
    # Reusable preprocessing
    transform = create_transform_component(preprocessing_config)
    
    # Configurable training
    trainer = create_trainer_component(training_config)
    
    return tfx.dsl.Pipeline(
        pipeline_name="modular-training-pipeline",
        components=[example_gen, transform, trainer]
    )

# ❌ Avoid: Hardcoded, monolithic pipelines
def create_pipeline():
    # Hardcoded paths and configurations
    example_gen = tfx.components.CsvExampleGen(
        input_base="/hardcoded/path/to/data"
    )
    # ... rest of pipeline with hardcoded values
```

#### 2. Proper Parameterization
**Use environment variables for configuration and runtime parameters for experimentation:**

```yaml
# Pipeline resource with proper parameterization
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Pipeline
metadata:
  name: training-pipeline
spec:
  image: "my-registry/ml-pipeline:v1.2.0"
  env:
    # Compile-time parameters (environment-specific)
    - name: DATA_ROOT
      value: "gs://production-bucket/data"
    - name: MODEL_REGISTRY
      value: "gs://model-registry/models"
    - name: PREPROCESSING_CONFIG
      value: "production"
```

```yaml
# Run with runtime parameters (experiment-specific)
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Run
metadata:
  name: experiment-lr-001
spec:
  provider: provider-namespace/provider-name
  pipeline: training-pipeline
  parameters:
    - name: learning_rate
      value: "0.001"
    - name: num_epochs
      value: "50"
    - name: batch_size
      value: "64"
    - name: dropout_rate
      value: "0.2"
```

#### 3. Comprehensive Testing Strategy
**Test pipelines at multiple levels:**

```python
# Unit tests for individual components
def test_preprocessing_component():
    # Test data transformation logic
    input_data = create_test_data()
    result = preprocess_data(input_data)
    assert result.shape == expected_shape
    assert result.columns == expected_columns

# Integration tests for pipeline compilation
def test_pipeline_compilation():
    pipeline = create_pipeline(
        data_root="gs://test-bucket/data",
        model_root="gs://test-bucket/models"
    )
    # Verify pipeline compiles without errors
    assert pipeline is not None
    assert len(pipeline.components) > 0

# End-to-end tests with test data
def test_pipeline_execution():
    # Run pipeline with small test dataset
    # Verify outputs are generated correctly
    pass
```

### Container Best Practices

#### 1. Efficient Docker Images
**Build optimized, secure container images:**

```dockerfile
# ✅ Good: Multi-stage build with security best practices
FROM python:3.9-slim as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Production stage
FROM python:3.9-slim

# Create non-root user
RUN useradd --create-home --shell /bin/bash mluser

# Copy installed packages from builder
COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy application code
COPY --chown=mluser:mluser pipeline/ /app/pipeline/
WORKDIR /app

# Switch to non-root user
USER mluser

# Set entrypoint
ENTRYPOINT ["python", "-m", "pipeline.main"]
```

#### 2. Image Versioning and Tagging
**Use semantic versioning and meaningful tags:**

```bash
# ✅ Good: Semantic versioning with descriptive tags
docker build -t my-registry/ml-pipeline:v1.2.0 .
docker tag my-registry/ml-pipeline:v1.2.0 my-registry/ml-pipeline:latest
docker tag my-registry/ml-pipeline:v1.2.0 my-registry/ml-pipeline:stable

# ✅ Good: Environment-specific tags
docker tag my-registry/ml-pipeline:v1.2.0 my-registry/ml-pipeline:production-v1.2.0

# ❌ Avoid: Generic or unclear tags
docker build -t my-registry/ml-pipeline:latest .
docker build -t my-registry/ml-pipeline:test .
```

## Resource Management

### Pipeline Resources

#### 1. Proper Resource Allocation
**Set appropriate resource requests and limits:**

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: training-pipeline
spec:
  provider: provider-namespace/provider-name
  image: "my-registry/ml-pipeline:v1.2.0"
  framework:
    name: tfx
    parameters:
      pipeline: training_pipeline.create_components
  env:
    - name: GPU_ENABLED
      value: "true"
    - name: NODE_SELECTOR
      value: "accelerator=nvidia-tesla-v100"
```

#### 2. Effective Labeling and Organization
**Use consistent labeling for resource management:**

```yaml
metadata:
  name: customer-churn-training
  labels:
    # Team and ownership
    team: "ml-engineering"
    owner: "data-science-team"
    
    # Project and domain
    project: "customer-churn"
    domain: "marketing"
    
    # Environment and lifecycle
    environment: "production"
    lifecycle: "active"
    
    # Version and release
    version: "v2.1.0"
    release: "2024-q1"
    
    # Cost tracking
    cost-center: "ml-infrastructure"
    budget: "ml-training"
```

### Namespace Strategy

#### 1. Environment Separation
**Use namespaces to separate environments:**

```bash
# Development environment
kubectl create namespace ml-dev
kubectl label namespace ml-dev environment=development

# Staging environment
kubectl create namespace ml-staging
kubectl label namespace ml-staging environment=staging

# Production environment
kubectl create namespace ml-prod
kubectl label namespace ml-prod environment=production
```

#### 2. Team-Based Organization
**Organize resources by team or project:**

```bash
# Team-based namespaces
kubectl create namespace ml-team-nlp
kubectl create namespace ml-team-vision
kubectl create namespace ml-team-recommendations

# Project-based namespaces
kubectl create namespace customer-churn
kubectl create namespace fraud-detection
kubectl create namespace recommendation-engine
```

## Automation and Scheduling

### RunConfiguration Best Practices

#### 1. Intelligent Scheduling
**Set up appropriate scheduling based on data availability and business needs:**

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: RunConfiguration
metadata:
  name: daily-model-training
spec:
  run:
    pipeline: training-pipeline
    runtimeParameters:
      data_date: "{{ .Date }}"
  schedule:
    cron: "0 6 * * *"  # 6 AM daily
```

## Security and Compliance

## Monitoring and Observability

### Pipeline Monitoring

#### 1. Comprehensive Logging
**Implement structured logging throughout your pipeline:**

```python
import logging
import json

# Configure structured logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

def log_pipeline_event(event_type, details):
    log_entry = {
        'event_type': event_type,
        'timestamp': datetime.utcnow().isoformat(),
        'pipeline_name': os.environ.get('PIPELINE_NAME'),
        'run_id': os.environ.get('RUN_ID'),
        'details': details
    }
    logger.info(json.dumps(log_entry))

# Usage in pipeline components
log_pipeline_event('data_validation_start', {
    'dataset_size': len(dataset),
    'validation_rules': validation_config
})
```

## Version Control

### Pipeline Versioning

#### 1. Semantic Versioning
**Version your pipelines semantically:**

```yaml
# Version pipeline resources
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: customer-churn-training
  labels:
    version: "v2.1.0"  # Major.Minor.Patch
    release: "2024-q1"
spec:
  provider: provider-namespace/provider-name
  image: "my-registry/customer-churn-pipeline:v2.1.0"
  framework:
    name: tfx
    parameters:
      pipeline: customer_churn_training.create_components
```