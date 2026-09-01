---
title: "TensorFlow Extended (TFX)"
linkTitle: "TFX"
type: docs
weight: 1
---

To create a TFX pipeline:
1) Ensure your [Provider](../providers/overview/) supports TFX by specifying the TFX image in `spec.frameworks[]`.
2) Create a [Pipeline resource](../resources/pipeline/), specifying:
- the `tfx` framework in `spec.framework.name`. This needs to match the name specified in the Provider.
- the fully qualified name of the Python function creating TFX pipeline components under `spec.framework.parameters[].components`.
- any required [beam arguments](https://www.tensorflow.org/tfx/guide/beam#beam_pipeline_arguments) under `spec.framework.parameters[].beamArgs`.

> [!NOTE]
>
> We aim to track current Python versions and follow Python's release lifecycle, dropping each version once it reaches [end of life](https://devguide.python.org/versions/). We advise using the newest version the compilers support (currently 3.12).

## TFX Parameters

| Name                  | Description                                                                                        |
|-----------------------|----------------------------------------------------------------------------------------------------|
| `components`          | Fully qualified name of the Python function creating TFX pipeline components.                      |
| `beamArgs[]`          | List of named objects. These will be provided as `beam_pipeline_args` when compiling the pipeline. |
| `usePipelineSpec2_1`  | Boolean. Use PipelineSpec 2.1 when `true` (default), or 2.0 when `false`.                          |


### TFX Pipeline resource example

{{% readfile file="/includes/master/quickstart/resources/pipeline.yaml" code="true" lang="yaml"%}}

## Component naming

The `COMPONENT` used in a Run's artifact `path` must match the component's id exactly:

- By default, the id is the component's class name.
- If the component overrides it via `.with_id("...")`, that string takes precedence — use it exactly.

| Pipeline code                       | COMPONENT to use |
|-------------------------------------|------------------|
| `Pusher(...)` (no `with_id`)        | `Pusher`         |
| `Pusher(...).with_id("push_model")` | `push_model`     |
