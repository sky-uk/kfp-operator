---
title: "Kubeflow Pipelines SDK (KFP SDK)"
linkTitle: "KFP SDK"
type: docs
weight: 2
---

To create a KFP SDK pipeline:
- Ensure your [Provider](../providers/overview/) supports KFP SDK by specifying the KFP SDK image in `spec.frameworks[]`.
- Create a [Pipeline resource](../resources/pipeline/), specifying:
  - the `kfpsdk` framework in `spec.framework.name`. This needs to match the name specified in the Provider.
  - the fully qualified name of the Python function that creates a KFP SDK pipeline under `spec.framework.parameters[].pipeline`.

## KFP SDK Parameters

| Name       | Description                                                                                                                                                                                                                               |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline` | Fully qualified name of the Python function creating a KFP SDK pipeline. This function should be wrapped using the [`kfp.dsl.Pipeline` decorator](https://kubeflow-pipelines.readthedocs.io/en/2.0.0b6/source/dsl.html#kfp.dsl.pipeline). |

### KFP SDK Pipeline resource example

{{% readfile file="/includes/master/kfpsdk-quickstart/resources/pipeline.yaml" code="true" lang="yaml"%}}
