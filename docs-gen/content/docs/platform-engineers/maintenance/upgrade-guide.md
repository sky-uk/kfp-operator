---
title: "Upgrade Guide"
linkTitle: "Upgrade Guide"
description: "Safe upgrade procedures and best practices for the KFP Operator platform"
weight: 10
---

# KFP Operator Upgrade Guide

This guide provides step-by-step procedures for safely upgrading the KFP Operator to new versions, including both stable and unstable CRD versions.

## Prerequisites

Before starting any upgrade, ensure you have:

- **Cluster Access**: Administrative access to your Kubernetes cluster
- **Helm Installed**: Helm 3.x installed and configured
- **Backup Strategy**: Current resource configurations backed up
- **Maintenance Window**: Scheduled downtime if required
- **Rollback Plan**: Clear rollback procedure documented

> **Important**: Always test upgrades in a non-production environment first.

---

## Standard Upgrade (Stable CRD Version)

Use this procedure when upgrading to a **stable, released version** of the KFP Operator.

### When to Use This Method

- Upgrading between stable releases (e.g., v0.7.0 → v0.8.0)
- Moving to a new stable CRD version (e.g., v1alpha6 → v1beta1)
- Production deployments requiring maximum stability

### Step-by-Step Procedure

#### Step 1: Configure Stored Version

**Purpose**: Ensure Kubernetes stores resources in the new CRD version format.

1. **Check your current Helm values file** (`values.yaml`):
   ```bash
   # View current configuration
   helm get values kfp-operator -n kfp-operator-system
   ```

2. **Update the stored version** in your `values.yaml`:
   ```yaml
   manager:
     multiversion:
       storedVersion: v1beta1  # Set to target version
   ```

3. **Verify the configuration** (see [Configuration Reference](../configuration/operator-configuration) for all options):
   ```bash
   # Validate your values file
   helm template kfp-operator ./helm-chart -f values.yaml --dry-run
   ```

> **Tip**: If you're using default Helm values without a custom `values.yaml`, you can skip this step as the stored version is automatically set to the latest stable version.

#### Step 2: Perform the Upgrade

**Purpose**: Deploy the new operator version with updated CRDs.

1. **Update your Helm repository**:
   ```bash
   helm repo update
   ```

2. **Upgrade the operator**:
   ```bash
   helm upgrade kfp-operator kfp-operator/kfp-operator \
     -n kfp-operator-system \
     -f values.yaml \
     --wait --timeout=10m
   ```

3. **Monitor the upgrade progress**:
   ```bash
   # Watch operator pods
   kubectl get pods -n kfp-operator-system -w

   # Check operator logs
   kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager -f
   ```

#### Step 3: Verify the Upgrade

**Purpose**: Confirm the upgrade completed successfully and resources are functioning.

1. **Check operator status**:
   ```bash
   # Verify operator is running
   kubectl get deployment -n kfp-operator-system

   # Check CRD versions
   kubectl get crd | grep pipelines.kubeflow.org
   ```

2. **Validate existing resources**:
   ```bash
   # List all pipeline resources
   kubectl get pipelines,runs,runconfigurations,providers --all-namespaces

   # Check resource status
   kubectl describe pipeline <pipeline-name> -n <namespace>
   ```

3. **Test basic functionality**:
   ```bash
   # Create a test pipeline (optional)
   kubectl apply -f - <<EOF
   apiVersion: pipelines.kubeflow.org/v1beta1
   kind: Pipeline
   metadata:
     name: upgrade-test
     namespace: default
   spec:
     image: "hello-world:latest"
   EOF
   ```

> **Success Indicators**:
> - All operator pods are `Running`
> - Existing resources show `Ready` status
> - No error messages in operator logs
> - Test resources can be created successfully

---

## Advanced Upgrade (Unstable CRD Version)

Use this procedure when upgrading to **development or pre-release versions** while maintaining easy rollback capability.

### When to Use This Method

- Testing latest features from the `master` branch
- Evaluating pre-release versions in staging environments
- Contributing to operator development and testing
- Gradual migration to new CRD versions

> **Warning**: Unstable versions are not recommended for production use. Always test thoroughly in non-production environments.

### Understanding the Strategy

This approach keeps the **stored version** on a stable release while allowing you to test new features:

- **Stored Version**: Remains on stable version (e.g., `v1alpha6`)
- **Served Version**: Includes both stable and unstable versions
- **Default Version**: Kubernetes serves the latest version to clients
- **Easy Rollback**: Since storage remains stable, rollback is straightforward

### Step-by-Step Procedure

#### Step 1: Configure for Unstable Version

**Purpose**: Set up the operator to test new features while maintaining stable storage.

1. **Identify current stable version**:
   ```bash
   # Check current CRD versions
   kubectl get crd pipelines.pipelines.kubeflow.org -o jsonpath='{.spec.versions[*].name}'

   # Find the current stored version
   kubectl get crd pipelines.pipelines.kubeflow.org -o jsonpath='{.spec.versions[?(@.storage==true)].name}'
   ```

2. **Create or update your `values.yaml`**:
   ```yaml
   manager:
     multiversion:
       storedVersion: v1alpha6  # Keep current stable version

   # Example: Using development image
   manager:
     image:
       repository: kfp-operator
       tag: "master-abc123"  # Development tag
   ```

3. **Validate configuration**:
   ```bash
   # Verify the stored version setting
   grep -A 5 "multiversion:" values.yaml
   ```

#### Step 2: Deploy Unstable Version

**Purpose**: Install the development version while maintaining stable storage.

1. **Backup current state** (recommended):
   ```bash
   # Export current resources
   kubectl get pipelines,runs,runconfigurations,providers --all-namespaces -o yaml > backup-before-unstable.yaml
   ```

2. **Perform the upgrade**:
   ```bash
   helm upgrade kfp-operator kfp-operator/kfp-operator \
     -n kfp-operator-system \
     -f values.yaml \
     --wait --timeout=15m
   ```

3. **Monitor deployment**:
   ```bash
   # Watch for any issues during deployment
   kubectl get events -n kfp-operator-system --sort-by='.lastTimestamp'
   ```

#### Step 3: Understand Version Behavior

**Purpose**: Know how Kubernetes handles multiple CRD versions.

**Version Priority Algorithm**: Kubernetes serves the highest priority version by default:
- Stable versions (v1, v1beta1) have higher priority than alpha versions
- Higher version numbers have priority (v1beta2 > v1beta1)
- Newer versions generally have higher priority

**Automatic Conversion**:
```bash
# Resources are automatically converted between versions
# Example: Resource stored as v1alpha6, served as v1beta1
kubectl get pipeline my-pipeline -o yaml
# Shows: apiVersion: pipelines.kubeflow.org/v1beta1 (latest served)

# But stored internally as v1alpha6 (stable storage)
```

**Conversion Webhooks**: Handle translation between versions seamlessly
- No data loss during conversion
- Bidirectional conversion support
- Automatic field mapping and transformation

#### Step 4: Verify Unstable Version

**Purpose**: Confirm the unstable version is working correctly.

1. **Check version serving**:
   ```bash
   # Verify both versions are served
   kubectl get crd pipelines.pipelines.kubeflow.org -o jsonpath='{.spec.versions[*].served}'

   # Confirm storage version unchanged
   kubectl get crd pipelines.pipelines.kubeflow.org -o jsonpath='{.spec.versions[?(@.storage==true)].name}'
   ```

2. **Test version conversion**:
   ```bash
   # Create resource with specific version
   kubectl apply -f - <<EOF
   apiVersion: pipelines.kubeflow.org/v1alpha6
   kind: Pipeline
   metadata:
     name: version-test
     namespace: default
   spec:
     image: "test:latest"
   EOF

   # Retrieve with latest version (should auto-convert)
   kubectl get pipeline version-test -o yaml | grep apiVersion
   ```

3. **Validate new features** (if applicable):
   ```bash
   # Test any new fields or functionality
   # Example: New field in v1beta1
   kubectl patch pipeline version-test --type='merge' -p='{"spec":{"newField":"value"}}'
   ```

#### Step 5: Incremental Updates

**Purpose**: Continue testing with newer unstable versions.

1. **Update to newer commits**:
   ```yaml
   # Update values.yaml with newer development tag
   manager:
     image:
       tag: "master-def456"  # Newer commit
     multiversion:
       storedVersion: v1alpha6  # Keep stable storage
   ```

2. **Apply incremental updates**:
   ```bash
   helm upgrade kfp-operator kfp-operator/kfp-operator \
     -n kfp-operator-system \
     -f values.yaml
   ```

#### Step 6: Promote to Stable (When Ready)

**Purpose**: Move to stable version once testing is complete.

1. **Update stored version**:
   ```yaml
   manager:
     multiversion:
       storedVersion: v1beta1  # Promote to stable
   ```

2. **Follow standard upgrade procedure**:
   - Use the [Standard Upgrade](#-standard-upgrade-stable-crd-version) process
   - This will migrate storage to the new stable version

### Rollback Procedures

#### Quick Rollback (Unstable Versions)

**When**: Rolling back from unstable version that was never set as stored version.

```bash
# Simple rollback to previous commit
helm rollback kfp-operator -n kfp-operator-system

# Or deploy specific stable version
helm upgrade kfp-operator kfp-operator/kfp-operator \
  -n kfp-operator-system \
  --version=0.7.0  # Specific stable version
```

> **Safe Rollback**: As long as the unstable version was never set as the stored version, rollback is guaranteed to work without data loss.

#### Emergency Rollback

**When**: Critical issues require immediate rollback.

1. **Immediate rollback**:
   ```bash
   # Rollback to last known good state
   helm rollback kfp-operator 1 -n kfp-operator-system
   ```

2. **Verify rollback**:
   ```bash
   # Check operator status
   kubectl get pods -n kfp-operator-system

   # Verify resources are accessible
   kubectl get pipelines --all-namespaces
   ```

3. **Restore from backup** (if needed):
   ```bash
   # Restore resources from backup
   kubectl apply -f backup-before-unstable.yaml
   ```

---

## Troubleshooting Common Issues

### Upgrade Stuck or Failing

**Symptoms**: Helm upgrade hangs or fails with timeout errors.

**Diagnosis**:
```bash
# Check operator pod status
kubectl get pods -n kfp-operator-system

# Check for resource conflicts
kubectl get events -n kfp-operator-system --sort-by='.lastTimestamp'

# Review operator logs
kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager --previous
```

**Solutions**:
1. **Increase timeout**:
   ```bash
   helm upgrade kfp-operator kfp-operator/kfp-operator \
     -n kfp-operator-system \
     -f values.yaml \
     --timeout=20m
   ```

2. **Force upgrade** (use with caution):
   ```bash
   helm upgrade kfp-operator kfp-operator/kfp-operator \
     -n kfp-operator-system \
     -f values.yaml \
     --force
   ```

### CRD Version Conflicts

**Symptoms**: Resources show incorrect versions or conversion errors.

**Diagnosis**:
```bash
# Check CRD versions and storage
kubectl get crd pipelines.pipelines.kubeflow.org -o yaml | grep -A 10 versions

# Verify conversion webhooks
kubectl get validatingwebhookconfiguration | grep kfp-operator
kubectl get mutatingwebhookconfiguration | grep kfp-operator
```

**Solutions**:
1. **Verify stored version configuration**:
   ```bash
   helm get values kfp-operator -n kfp-operator-system
   ```

2. **Manually update CRD** (advanced):
   ```bash
   # Only if automatic update fails
   kubectl apply -f https://raw.githubusercontent.com/sky-uk/kfp-operator/main/config/crd/bases/
   ```

### Resource Reconciliation Issues

**Symptoms**: Existing resources show `Failed` or `Unknown` status after upgrade.

**Diagnosis**:
```bash
# Check resource status
kubectl get pipelines,runs,runconfigurations,providers --all-namespaces

# Describe problematic resources
kubectl describe pipeline <name> -n <namespace>

# Check controller logs
kubectl logs -n kfp-operator-system deployment/kfp-operator-controller-manager | grep ERROR
```

**Solutions**:
1. **Restart operator**:
   ```bash
   kubectl rollout restart deployment/kfp-operator-controller-manager -n kfp-operator-system
   ```

2. **Re-apply resources**:
   ```bash
   # Export and re-apply resource
   kubectl get pipeline <name> -n <namespace> -o yaml > pipeline-backup.yaml
   kubectl delete pipeline <name> -n <namespace>
   kubectl apply -f pipeline-backup.yaml
   ```

### Webhook Certificate Issues

**Symptoms**: Admission webhook errors or certificate validation failures.

**Diagnosis**:
```bash
# Check webhook configurations
kubectl get validatingwebhookconfiguration kfp-operator-validating-webhook-configuration -o yaml

# Check certificate secrets
kubectl get secret -n kfp-operator-system | grep webhook
```

**Solutions**:
1. **Regenerate certificates**:
   ```bash
   # Delete webhook certificates (they will be regenerated)
   kubectl delete secret webhook-server-certs -n kfp-operator-system

   # Restart operator to regenerate
   kubectl rollout restart deployment/kfp-operator-controller-manager -n kfp-operator-system
   ```

---

## Best Practices

### Pre-Upgrade Checklist

- [ ] **Backup Resources**: Export all custom resources to YAML files
- [ ] **Document Current State**: Record current operator version and CRD versions
- [ ] **Test in Staging**: Perform upgrade in non-production environment first
- [ ] **Check Dependencies**: Verify compatibility with other cluster components
- [ ] **Plan Maintenance Window**: Schedule appropriate downtime if needed
- [ ] **Prepare Rollback Plan**: Document exact rollback procedures
- [ ] **Monitor Resources**: Set up monitoring for upgrade process

### During Upgrade

- [ ] **Monitor Logs**: Watch operator logs for errors or warnings
- [ ] **Check Resource Status**: Verify existing resources remain healthy
- [ ] **Validate Functionality**: Test basic operations after upgrade
- [ ] **Document Issues**: Record any problems encountered for future reference

### Post-Upgrade

- [ ] **Verify All Resources**: Confirm all pipelines, runs, and configurations are working
- [ ] **Test New Features**: Validate any new functionality introduced
- [ ] **Update Documentation**: Record successful upgrade and any lessons learned
- [ ] **Clean Up**: Remove backup files and temporary resources if upgrade successful
- [ ] **Monitor Performance**: Watch for any performance impacts over time

### Version Management Strategy

#### For Production Environments
```yaml
# Recommended: Always use stable versions
manager:
  multiversion:
    storedVersion: v1beta1  # Latest stable
  image:
    tag: "v0.8.0"  # Stable release tag
```

#### For Development/Testing
```yaml
# Acceptable: Test unstable versions
manager:
  multiversion:
    storedVersion: v1alpha6  # Keep stable storage
  image:
    tag: "master-abc123"  # Development tag
```

#### Version Progression Path
1. **Alpha** (`v1alpha1`, `v1alpha2`, etc.) → Development and early testing
2. **Beta** (`v1beta1`, `v1beta2`, etc.) → Feature complete, API stable
3. **Stable** (`v1`, `v2`, etc.) → Production ready, long-term support

---

## Additional Resources

### Configuration References
- **[Operator Configuration](../configuration/operator-configuration)**: Complete configuration options
- **[Provider Configuration](../configuration/providers/)**: Provider-specific settings
- **[Helm Chart Values](https://github.com/sky-uk/kfp-operator/tree/main/helm-chart)**: Default Helm values

### Kubernetes Documentation
- **[CRD Versioning](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)**: Kubernetes CRD version concepts
- **[Version Priority](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#version-priority)**: How Kubernetes prioritizes versions
- **[Conversion Webhooks](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#webhook-conversion)**: Automatic version conversion

### Support and Community
- **[GitHub Issues](https://github.com/sky-uk/kfp-operator/issues)**: Report bugs or request features
- **[GitHub Discussions](https://github.com/sky-uk/kfp-operator/discussions)**: Community support and questions
- **[Release Notes](https://github.com/sky-uk/kfp-operator/releases)**: Detailed changelog for each version

---

## Emergency Contacts

If you encounter critical issues during upgrade:

1. **Immediate Rollback**: Use the rollback procedures above
2. **Check Documentation**: Review troubleshooting section
3. **Community Support**: Post in GitHub Discussions with detailed error information
4. **Emergency Issues**: Create GitHub issue with `urgent` label

> **Remember**: When in doubt, rollback first, then investigate. The stored version strategy ensures safe rollbacks are always possible.
