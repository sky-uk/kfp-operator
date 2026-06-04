# TFX Patch Image

A lightweight `scratch`-based image containing a single Python script (`patch_tfx_kfp_v2.py`) that fixes five TFX runtime incompatibilities with the OSS KFP v2 driver/launcher.

## Patches

| # | Target | Fix |
|---|--------|-----|
| 1 | `compiler_utils.py` | Map `system.*` schema titles to TFX artifact classes |
| 2 | `kubeflow_v2_entrypoint_utils.py` | Handle `None` `instance_schema` from the KFP v2 driver |
| 3 | `kubeflow_v2_entrypoint_utils.py` | Infer correct TFX artifact class from metadata keys |
| 4 | `kubeflow_v2_run_executor.py` | Enrich input artifact types from component inputs spec |
| 5 | `path_utils.py` | Flatten model directory to work around KFP launcher bug |

Patches 1–4 are safe for all environments. Patch 5 changes the TFX model layout and is incompatible with Vertex AI Pipelines.
There is an open issue + PR in Kubeflow to fix the launcher bug instead (Patch 5). Which will make this redundant, but until then it is required.

## Supported TFX versions

1.15, 1.16, 1.17 — the patched files are identical across all three versions.

## Usage

Pipeline authors consume the patch via a Docker multi-stage build: e.g

> [!NOTE]
> The patch must be applied after the TFX image is installed / uv sync'd to take effect.

```dockerfile
FROM ghcr.io/kfp-operator/kfp-operator-tfx-patch:latest AS tfx-patch
FROM tensorflow/tfx:1.17.3

COPY --from=tfx-patch /patch_tfx_kfp_v2.py /tmp/
RUN python /tmp/patch_tfx_kfp_v2.py && rm /tmp/patch_tfx_kfp_v2.py

COPY my_pipeline/ /pipeline/
```

To skip patch 5 for images that also run on Vertex AI:

```dockerfile
RUN python /tmp/patch_tfx_kfp_v2.py --vertex-compatible && rm /tmp/patch_tfx_kfp_v2.py
```

You can also verify the patch can be applied without modifying files by running with `--check`:

```dockerfile
RUN python /tmp/patch_tfx_kfp_v2.py --check && rm /tmp/patch_tfx_kfp_v2.py
```

## Build

```bash
make docker-build
```
