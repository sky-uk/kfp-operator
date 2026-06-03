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

## Supported TFX versions

1.15, 1.16, 1.17 — the patched files are identical across all three versions.

## Usage

Pipeline authors consume the patch via a Docker multi-stage build:

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

## Build

```bash
make docker-build
```
