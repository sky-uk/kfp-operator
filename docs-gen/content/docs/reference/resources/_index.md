---
title: "Resources"
weight: 2
---

The Kubeflow Pipelines operator manages the lifecycle of pipelines and related resources via Kubernetes Resources.

All resources managed by the operator have the following common status fields:

| Name | Description |
| --- | --- |
| `providerId` | The resource identifier inside Kubeflow Pipelines |
| `version` | The resource version |
| `synchronizationState` | The current synchronization state with Kubeflow Pipelines |
| `observedGeneration` | The last processed [generation](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) of the resource |
