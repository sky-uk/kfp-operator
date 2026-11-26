# TFX Patch Image

A `scratch`-based image containing the `tfx_kfp_v2_shim` package. It installs a PEP 451 import hook that patches TFX at runtime for OSS KFP v2 compatibility.

> [!Important]
> Only tested against the quickstart image in this repo. Not guaranteed to work with all TFX pipelines.

## How it works

`install_shim.py` copies the package into site-packages and writes a `.pth` file so the hook activates on every Python invocation. When a target TFX module is imported the hook lets the real loader run, then monkey-patches the module in place.

## Patches

| Patch | Target module | What it does |
|-------|---------------|--------------|
| Artifact type mappings (`patch_compiler_utils`) | `compiler_utils` | Add `system.*` → TFX class mappings to `TITLE_TO_CLASS_PATH` |
| Null-schema handling (`patch_entrypoint_utils`) | `kubeflow_v2_entrypoint_utils` | Guard `None` `instance_schema` (`yaml.safe_load("")` → `None`) |
| Type inference (`patch_entrypoint_utils`) | `kubeflow_v2_entrypoint_utils` | Infer correct TFX type from artifact metadata keys |
| Input artifact schema enrichment (`patch_entrypoint_utils`) | `kubeflow_v2_entrypoint_utils` | Copy artifact type schemas from `inputs_spec` into untyped artifacts |
| GCS directory-marker cleanup (`patch_entrypoint_utils`) | `kubeflow_v2_entrypoint_utils` | Delete zero-byte GCS directory markers after each executor run (`gs://` outputs only) |
| DAG runner compatibility (`patch_experimental`) | `tfx.v1.orchestration.experimental` | Re-export `KubeflowV2DagRunner` removed by `kfp>=2` |
| Executor shutdown compatibility (`patch_run_executor`) | `kubeflow_v2_run_executor` | Force-exit(0) after output metadata is written, to bypass a C++ destructor crash ("pure virtual method called") during interpreter shutdown |

The same image works on Vertex AI and KFP, with GCS or S3 artifact storage. No patch changes artifact layout; models are still written under `Format-Serving/`.

Executor shutdown compatibility is applied as a **source edit** at install time (`install_shim.patch_run_executor_source`), not via the runtime hook: the executor container runs `python -m ...kubeflow_v2_run_executor`, i.e. as `__main__`, which the import hook cannot patch (runpy uses the loader's `get_code`, bypassing `exec_module`).

GCS directory-marker cleanup works around a KFP launcher bug ([kubeflow/pipelines#13476](https://github.com/kubeflow/pipelines/issues/13476)): the launcher downloads the zero-byte directory markers written by TensorFlow's GCS filesystem as files, breaking downstream artifact downloads. It only matters on KFP with GCS, is harmless on Vertex AI, and does nothing for other storage such as S3. If the GCS client or credentials are unavailable, the cleanup is skipped with a warning. It becomes redundant once the bug is fixed upstream.

## Supported TFX versions

1.15, 1.16, 1.17

## Usage

Source the shim directly via a named build context so the image is
self-contained (no prebuilt patch image to pull). Point the `tfx-shim` context
at this `tfx_kfp_v2_shim` directory when building:

```bash
docker build --build-context tfx-shim=path/to/provider-service/kfp/tfx-patch/tfx_kfp_v2_shim ...
```

Then install it in a multi-stage build after TFX dependencies are in place:

```dockerfile
FROM python:3.10.19

# install TFX and dependencies first
COPY --from=tfx-shim . /tmp/tfx_kfp_v2_shim/
RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim
```

## Testing

```bash
make test
```

## Build

```bash
make docker-build
```
