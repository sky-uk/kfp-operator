---
title: "Kubeflow Pipelines V2"
---

## Overview

The **Kubeflow Pipelines Provider Service** is included within this project to interface directly with the
Kubeflow Pipelines V2 backend. This service acts as a bridge between the KFP Operator and Kubeflow Pipelines V2,
facilitating operations such as pipeline submission, status monitoring, schedules and experiment management.

![KFP Provider]({{< param "subpath" >}}/master/images/kfp.svg)

## Deployment and Usage

Set up the service with the necessary configurations, including API endpoints and authentication
credentials for the Kubeflow Pipelines Provider. [See the getting started guide.](../../../getting-started/installation/#providers)

Kubeflow Pipelines must be [installed](https://www.kubeflow.org/docs/components/pipelines/operator-guides/installation/).
Its configuration can be controlled using the [KFP specific parameters within a Provider Resource](../../resources/provider/#kubeflow).

For detailed implementation code and further technical insights, refer to the
[KFP Provider Service directory](https://github.com/sky-uk/kfp-operator/tree/master/provider-service/kfp) in the
official repository.

## Implementation Details

- **API** : Implements the [openapi spec for provider services](../overview/#api).

- **Event Handling**: The KFP provider creates run-completion events when reading the status of workflows triggered by
Kubeflow. These events are then processed and sent to the operator's webhook to update the status of the run.

## Platform Settings

The Provider's `parameters.platform` holds settings the provider adds to pipeline components each time you create or
update a pipeline. The provider service refuses to start if they are invalid. When you change them, the operator
redeploys the provider service.

### Accelerator Placement

Pipeline components can request an accelerator, such as a GPU, by type. You map each type to the Kubernetes resource
the provider sets as the container's limit, and to the node selector and tolerations that schedule the component onto
nodes with that accelerator. Pipeline authors request a type, and you keep the cluster's node labels and taints on the
Provider.

List the types under the Provider's `parameters.platform.accelerators`:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: kfp
  namespace: kfp-operator
spec:
  parameters:
    platform:
      accelerators:
        nvidia-t4:
          resourceType: nvidia.com/gpu
          nodeSelector:
            nvidia.com/gpu-type: t4
          tolerations:
          - key: nvidia.com/gpu
            operator: Exists
            effect: NoSchedule
        nvidia-l4:
          resourceType: nvidia.com/gpu
          nodeSelector:
            nvidia.com/gpu-type: l4
          tolerations:
          - key: nvidia.com/gpu
            operator: Exists
            effect: NoSchedule
    # ...
```

| Field                                        | Description                                                                                                                     |
|----------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| `platform.accelerators.<type>`               | The accelerator type a pipeline component requests                                                                              |
| `platform.accelerators.<type>.resourceType`  | Required. The Kubernetes resource the provider sets as the container's limit, e.g. `nvidia.com/gpu`                             |
| `platform.accelerators.<type>.nodeSelector`  | Labels of the nodes Kubernetes schedules the component onto                                                                     |
| `platform.accelerators.<type>.tolerations[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) the provider adds to the component |

Each time you create or update a pipeline, the provider changes the components that request an accelerator:

- It replaces the requested type with the type's `resourceType` and keeps the count.
- It adds the `nodeSelector` unless the component sets its own.
- It adds any `tolerations` missing from the component.

The provider rejects the upload and keeps the pipeline's existing versions if a component requests a type that isn't
configured, or a type that a pipeline parameter sets at run time. A component that sets its own node selector may
request any type. The provider leaves components without an accelerator request unchanged, and without
`platform.accelerators` it uploads pipelines as compiled.

See the [TFX]({{< ref "/docs/ml-engineers/frameworks/tfx.md#requesting-accelerators" >}}) and
[KFP SDK]({{< ref "/docs/ml-engineers/frameworks/kfpsdk.md#requesting-accelerators" >}}) guides for how pipeline
authors request an accelerator.

### Executor Settings

The other keys under `parameters.platform` are Kubeflow Pipelines'
[Kubernetes executor settings](https://github.com/kubeflow/pipelines/blob/master/kubernetes_platform/proto/kubernetes_executor_config.proto),
which the KFP SDK's `kfp.kubernetes` module sets per component. The provider adds them to every component. For example,
to give every component cloud credentials from a Secret, with the path to the key file in an environment variable:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: google-credentials-env
  namespace: kubeflow
data:
  GOOGLE_APPLICATION_CREDENTIALS: /var/secrets/google/key.json
---
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: kfp
  namespace: kfp-operator
spec:
  parameters:
    platform:
      secretAsVolume:
      - secretName: google-credentials
        mountPath: /var/secrets/google
      configMapAsEnv:
      - configMapName: google-credentials-env
        keyToEnv:
        - configMapKey: GOOGLE_APPLICATION_CREDENTIALS
          envVar: GOOGLE_APPLICATION_CREDENTIALS
    # ...
```

The Secret and ConfigMap must exist in the namespace where pipeline components run.

A component keeps any setting it already has. List settings, such as `secretAsVolume` and `tolerations`, gain the items
the component lacks. The provider places accelerators first, so an accelerator's node selector takes precedence over a
`nodeSelector` set here. The provider supports the settings of the Kubeflow Pipelines version it is built against and
refuses to start on any other key.
