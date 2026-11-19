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

## Accelerator Placement

Pipeline components can request an accelerator, such as a GPU, by type. The provider can turn those types into
Kubernetes scheduling: each type maps to the Kubernetes resource the container is limited on, and the node selector and
tolerations that place the component on nodes with that accelerator. Pipeline authors then only request a type, and
the cluster's node labels and taints stay in one place.

List the types in a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kfp-accelerators
  namespace: kfp-operator
data:
  accelerators.yaml: |
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
```

| Field                                | Description                                                                                                       |
|--------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| `accelerators.<type>`                | The accelerator type a pipeline component requests                                                                |
| `accelerators.<type>.resourceType`   | Required. The Kubernetes resource the component's container is limited on, e.g. `nvidia.com/gpu`                  |
| `accelerators.<type>.nodeSelector`   | Node labels the component must be scheduled on                                                                     |
| `accelerators.<type>.tolerations[]`  | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) added to the component |

Mount the ConfigMap into the provider service and set `parameters.acceleratorConfigPath` to the file:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: kfp
  namespace: kfp-operator
spec:
  parameters:
    acceleratorConfigPath: /etc/kfp-provider/accelerators.yaml
    # ...
  podTemplateVolumes:
  - name: accelerators
    configMap:
      name: kfp-accelerators
  podTemplateVolumeMounts:
  - name: accelerators
    mountPath: /etc/kfp-provider
    readOnly: true
```

The provider reads the file when it starts and fails to start if the file is invalid, so restart the provider service
after changing the ConfigMap.

When a pipeline is created or updated, the provider changes every component that requests an accelerator:

- The requested type is replaced with its `resourceType`. The count is unchanged.
- The `nodeSelector` is added, unless the component already sets a node selector, in which case that is kept.
- The `tolerations` are added to any the component already has.

The upload fails, and the pipeline's existing versions are kept, if a component requests a type that is not
configured, or a type that is only resolved at run time from a pipeline parameter. A component that sets its own node
selector may request any type. Components that do not request an accelerator are not changed. Without
`acceleratorConfigPath`, pipelines are uploaded as compiled.

See the [TFX]({{< ref "/docs/ml-engineers/frameworks/tfx.md#requesting-accelerators" >}}) and
[KFP SDK]({{< ref "/docs/ml-engineers/frameworks/kfpsdk.md#requesting-accelerators" >}}) guides for how pipeline
authors request an accelerator.
