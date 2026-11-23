---
title: "Kubeflow Pipelines SDK (KFP SDK)"
linkTitle: "KFP SDK"
type: docs
weight: 2
---

To create a KFP SDK pipeline:
- Ensure your [Provider](../../../platform-engineers/configuration/providers/) supports KFP SDK by specifying the KFP SDK image in `spec.frameworks[]`.
- Create a [Pipeline resource](../resources/pipeline/), specifying:
  - the `kfpsdk` framework in `spec.framework.name`. This needs to match the name specified in the Provider.
  - the fully qualified name of the Python function that creates a KFP SDK pipeline under `spec.framework.parameters[].pipeline`.

> [!NOTE]
>
> We aim to track current Python versions and follow Python's release lifecycle, dropping each version once it reaches [end of life](https://devguide.python.org/versions/). We advise using the newest version the compilers support (currently 3.12).

## KFP SDK Parameters

| Name       | Description                                                                                                                                                                                                                               |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline` | Fully qualified name of the Python function creating a KFP SDK pipeline. This function should be wrapped using the [`kfp.dsl.Pipeline` decorator](https://kubeflow-pipelines.readthedocs.io/en/2.0.0b6/source/dsl.html#kfp.dsl.pipeline). |

## Environment Variables

The KFP SDK compiler automatically injects the following environment variables during compilation:

| Variable             | Description                                                                                                                                                                                                   |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `KFP_PIPELINE_IMAGE` | Set to the value of `spec.image` from the Pipeline resource. Use this in your pipeline code to dynamically set component base images, e.g. `@dsl.component(base_image=os.environ.get("KFP_PIPELINE_IMAGE"))`. |

## KFP SDK Pipeline resource example

{{% readfile file="/includes/master/kfpsdk-quickstart/resources/pipeline.yaml" code="true" lang="yaml"%}}

## Requesting accelerators

To request an accelerator, such as a GPU, call `set_accelerator_type` and `set_accelerator_limit` on the task:

```python
train(epochs=epochs).set_accelerator_type("nvidia-t4").set_accelerator_limit(1)
```

On Kubeflow Pipelines, use one of the accelerator types
[configured on the Provider]({{< ref "/docs/platform-engineers/configuration/providers/kfp.md#accelerator-placement" >}}).
The provider schedules the task onto matching nodes. If the type is not configured on your platform, the Pipeline
fails to create or update. A task that sets its own node selector with `kfp.kubernetes.add_node_selector` may request
any type.

## Component naming

The `COMPONENT` used in a Run's artifact `path` is derived from the `@dsl.component` Python function name, which the
KFP SDK normalizes: the name is lower-cased and every run of non-alphanumeric characters (e.g. underscores) is
replaced with a single hyphen. CamelCase is **not** split into words (e.g. a function named `BatchPredictor` becomes
`batchpredictor`, not `batch-predictor`).

| Pipeline code (function name) | COMPONENT to use |
|-------------------------------|------------------|
| `def pusher(...)`             | `pusher`         |
| `def train_model(...)`        | `train-model`    |
| `def BatchPredictor(...)`     | `batchpredictor` |
