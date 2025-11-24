---
title: "Troubleshooting"
linkTitle: "Troubleshooting"
description: "Debug and resolve common ML pipeline issues with the KFP Operator"
weight: 60
---

# Troubleshooting ML Pipelines

Diagnose and resolve common issues when developing and running ML pipelines with the KFP Operator.

## Quick Problem Resolution

### Most Common Issues

#### 1. Pipeline Won't Deploy
**Symptoms**: Pipeline resource shows error status
```bash
kubectl get pipeline my-pipeline
# NAME         STATUS   PROVIDER   AGE
# my-pipeline  Error    -          2m
```

**Quick Fixes**:
- Check image exists and is accessible
- Verify provider is ready and configured
- Check resource quotas and limits

**Detailed Guide**: [Pipeline Deployment Issues](#pipeline-deployment-issues)

#### 2. Run Fails to Start
**Symptoms**: Run resource stuck in pending state
```bash
kubectl get run my-run
# NAME     STATUS    PIPELINE     AGE
# my-run   Pending   my-pipeline  5m
```

**Quick Fixes**:
- Verify pipeline is ready
- Check runtime parameters are valid
- Ensure sufficient cluster resources

**Detailed Guide**: [Run Execution Issues](#run-execution-issues)

#### 3. Pipeline Components Fail
**Symptoms**: Individual pipeline steps fail during execution
```bash
kubectl logs -l workflows.argoproj.io/workflow=my-run
# Error: Component 'trainer' failed with exit code 1
```

**Quick Fixes**:
- Check component logs for specific errors
- Verify data inputs and outputs
- Check resource allocation

**Detailed Guide**: [Component Debugging](#component-debugging)

## Diagnostic Tools

### Essential Commands

#### Check Pipeline Status
```bash
# Get pipeline overview
kubectl get pipelines

# Detailed pipeline information
kubectl describe pipeline <pipeline-name>

# Check pipeline events
kubectl get events --field-selector involvedObject.name=<pipeline-name>
```

#### View Kubernetes Events
The operator emits Kubernetes events for all resource transitions which can be viewed using `kubectl describe`.

```bash
# View events for a specific pipeline
kubectl describe pipeline pipeline-sample

# Example output:
# Events:
#   Type     Reason      Age    From          Message
#   ----     ------      ----   ----          -------
#   Normal   Syncing     5m54s  kfp-operator  Updating [version: "v5-841641"]
#   Warning  SyncFailed  101s   kfp-operator  Failed [version: "v5-841641"]: pipeline update failed
#   Normal   Syncing     9m47s  kfp-operator  Updating [version: "57be7f4-681dd8"]
#   Normal   Synced      78s    kfp-operator  Succeeded [version: "57be7f4-681dd8"]
```

#### Monitor Run Execution
```bash
# Watch run progress
kubectl get run <run-name> -w

# Get run details
kubectl describe run <run-name>

# Check associated workflow
kubectl get workflows -l pipelines.kubeflow.org/run=<run-name>
```

#### View Logs
```bash
# Get workflow logs
kubectl logs -l workflows.argoproj.io/workflow=<run-name>

# Get specific component logs
kubectl logs <pod-name> -c <container-name>

# Follow logs in real-time
kubectl logs -f -l workflows.argoproj.io/workflow=<run-name>
```

### Debugging Workflow

1. **Check Resource Status**
   ```bash
   kubectl get pipeline,run,provider
   ```

2. **Examine Events**
   ```bash
   kubectl get events --sort-by='.lastTimestamp'
   ```

3. **Review Logs**
   ```bash
   kubectl logs -l app.kubernetes.io/name=kfp-operator
   ```

4. **Validate Configuration**
   ```bash
   kubectl get <resource> -o yaml
   ```

## Common Problems and Solutions

### Pipeline Deployment Issues

#### Problem: Image Pull Errors
**Error Message**: `Failed to pull image: unauthorized`

**Causes**:
- Container registry authentication issues
- Image doesn't exist or wrong tag
- Network connectivity problems

**Solutions**:
```bash
# Check if image exists
docker pull <your-image>

# Verify registry credentials
kubectl get secrets -o yaml | grep dockerconfigjson

# Test network connectivity
kubectl run test-pod --image=busybox --rm -it -- wget <registry-url>
```

#### Problem: Resource Validation Errors
**Error Message**: `Pipeline.spec.image is required`

**Causes**:
- Missing required fields in Pipeline spec
- Invalid YAML syntax
- Incorrect API version

**Solutions**:
```bash
# Validate YAML syntax
kubectl apply --dry-run=client -f pipeline.yaml

# Check API version
kubectl api-resources | grep pipelines

# Use kubectl explain for field reference
kubectl explain pipeline.spec
```

### Run Execution Issues

#### Problem: Parameter Validation Errors
**Error Message**: `Invalid runtime parameter: learning_rate`

**Causes**:
- Parameter not defined in pipeline
- Wrong parameter type or format
- Typo in parameter name

**Solutions**:
```python
# Define parameters in pipeline code
learning_rate = RuntimeParameter(
    name='learning_rate',
    default=0.001,
    ptype=float
)

# Use correct parameter in Run
spec:
  parameters:
    - name: learning_rate
      value: "0.001"  # String value, will be converted
```

### Enable Debug Logging

#### Pipeline-Level Debugging
```python
# Add debug logging to pipeline components
import logging
logging.basicConfig(level=logging.DEBUG)

def debug_component_execution():
    logger = logging.getLogger(__name__)
    logger.debug("Component starting execution")
    logger.debug(f"Input parameters: {locals()}")
```

#### Operator-Level Debugging
```bash
# Check operator logs with increased verbosity
kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager

# Enable debug mode (platform team may need to do this)
# Set logging.verbosity: 2 in operator configuration
```

### Pipeline Compilation Errors

### Compiling locally

The KFP-Operator's compiler can be used locally for debugging purposes. This can be especially useful for troubleshooting environment variable and beam argument resolution.

#### Environment setup and compiler injection

The compiler is injected into a shared directory first before it can be called from within the pipeline image.
Note that the setup is usually only needed once unless you want to use a different version of the compiler.

```shell
export KFP_COMPILER_IMAGE=ghcr.io/kfp-operator/kfp-operator-tfx-compiler:<KFP-Operator version>
docker pull $KFP_COMPILER_IMAGE

# Create a temporary directory for the following steps, alternatively choose a different location
SHARED_DIR=$(mktemp -d)

# Inject the compiler into the shared temporary directory
docker run -v $SHARED_DIR:/shared $KFP_COMPILER_IMAGE /shared
```

#### Compiler configuration

The compilation process relies on the pipeline resource and the provider configuration being passed:

```shell
export PIPELINE_IMAGE=<your pipeline image>

# create the pipeline resource
cat > $SHARED_DIR/pipeline.yaml << EOF
name: <Your pipeline name>
image: $PIPELINE_IMAGE
framework:
  name: tfx
  parameters:
    components: <component function>
    beamArgs:
      - [] # List of NamedValues for beam arguments
env:
  <Dict[str, str] of environment variables to be passed to the compilation step>
EOF
```

#### Running the compiler

You can then run the compiler from inside your pipeline container to produce `$SHARED_DIR/pipeline_out.yaml`.

```shell
# Run the compiler in your pipeline image
docker run -v $SHARED_DIR:/shared --entrypoint /shared/compile.sh $PIPELINE_IMAGE --pipeline_config /shared/pipeline.yaml --output_file /shared/pipeline_out.yaml
```


## Getting Additional Help

### When to Escalate

Contact your platform team if you encounter:
- **Operator installation issues**
- **Provider connectivity problems**
- **Cluster resource constraints**
- **RBAC or permission errors**

### Self-Service Debugging

You can resolve these issues yourself:
- **Pipeline code errors**
- **Runtime parameter issues**
- **Data processing problems**
- **Component configuration errors**

### Community Resources

- **[GitHub Issues](https://github.com/sky-uk/kfp-operator/issues)**: Search for known issues
- **[GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)**: Ask questions
- **[Platform Engineer Docs](../../platform-engineers/troubleshooting/)**: For platform-level issues

## Related Documentation

- **[Best Practices](../best-practices/)**: Prevent common issues
- **[API Reference](../../reference/)**: Understand resource specifications
- **[Tutorials](../getting-started/tutorials/)**: Learn through examples

---

**Still having issues?** Check the [detailed debugging guide](debugging/) for comprehensive troubleshooting procedures and advanced diagnostic techniques.
