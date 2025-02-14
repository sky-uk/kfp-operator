---
title: "Resources"
weight: 2
---

The Kubeflow Pipelines operator manages the lifecycle of pipelines and related resources via Kubernetes Resources.

All resources managed by the operator have the following common status fields:

| Name                   | Description                                                                                                                                     |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `synchronizationState` | The current synchronization state with the targeted provider                                                                                    |
| `observedGeneration`   | The last processed [generation](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) of the resource |

Additionally, all resources that are directly synchronised with a provider have the following status fields:

| Name         | Description                                          |
| ------------ | ---------------------------------------------------- |
| `providerId` | The resource identifier inside the targeted provider |
| `version`    | The resource version                                 |
