# TFX Patch Image

A `scratch`-based image containing the `tfx_kfp_v2_shim` package. It installs a PEP 451 import hook that patches TFX at runtime for OSS KFP v2 compatibility.

> [!Important]
> Only tested against the quickstart image in this repo. Not guaranteed to work with all TFX pipelines.

## How it works

`install_shim.py` copies the package into site-packages and writes a `.pth` file so the hook activates on every Python invocation. When a target TFX module is imported the hook lets the real loader run, then monkey-patches the module in place.

## Patches

| # | Target module | What it does |
|---|---------------|--------------|
| 1 | `compiler_utils` | Add `system.*` → TFX class mappings to `TITLE_TO_CLASS_PATH` |
| 2 | `kubeflow_v2_entrypoint_utils` | Guard `None` `instance_schema` (`yaml.safe_load("")` → `None`) |
| 3 | `kubeflow_v2_entrypoint_utils` | Infer correct TFX type from artifact metadata keys |
| 4 | `kubeflow_v2_entrypoint_utils` | Copy artifact type schemas from `inputs_spec` into untyped artifacts |
| 5 | `path_utils` | Flatten model directory (skip `Format-Serving/` subdirectory) |
| 6 | `tfx.v1.orchestration.experimental` | Re-export `KubeflowV2DagRunner` removed by `kfp>=2` |
| 7 | `kubeflow_v2_entrypoint_utils` | Delete zero-byte GCS directory markers after each executor run |
| 8 | `kubeflow_v2_run_executor` | Force-exit(0) after output metadata is written, to bypass a C++ destructor crash ("pure virtual method called") during interpreter shutdown |

Patches 1–4 and 6–8 are safe for all environments. Patch 5 changes TFX's model layout and breaks Vertex AI — use `--vertex-compatible` to skip it.

Patch 8 is applied as a **source edit** at install time (`install_shim.patch_run_executor_source`), not via the runtime hook: the executor container runs `python -m ...kubeflow_v2_run_executor`, i.e. as `__main__`, which the import hook cannot patch (runpy uses the loader's `get_code`, bypassing `exec_module`).

Patches 5 and 7 work around a KFP launcher bug ([kubeflow/pipelines#13476](https://github.com/kubeflow/pipelines/issues/13476)). They become redundant once that is fixed upstream.

## Supported TFX versions

1.15, 1.16, 1.17

## Usage

Install via multi-stage build after TFX dependencies are in place:

```dockerfile
FROM ghcr.io/kfp-operator/kfp-operator-tfx-patch:latest AS tfx-patch
FROM python:3.10.12

# install TFX and dependencies first
COPY --from=tfx-patch /tfx_kfp_v2_shim /tmp/tfx_kfp_v2_shim
RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim
```

Skip patch 5 for Vertex AI compatibility:

```dockerfile
RUN python /tmp/tfx_kfp_v2_shim/install_shim.py --vertex-compatible && rm -rf /tmp/tfx_kfp_v2_shim
```

## Testing

```bash
make test
```

## Build

```bash
make docker-build
```
