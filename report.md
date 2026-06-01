# TFX 1.17 + KFP v2 Compatibility Fixes Report

## Overview

Running TFX 1.17 pipelines on KFP v2 (via the kfp-operator) requires several patches to the TFX container image. These address incompatibilities between TFX's expected runtime environment and how KFP v2's driver/launcher manages artifacts and type metadata.

---

## Fix 1: Schema Title Mapping (`compiler_utils`)

**What**: Append additional entries to `TITLE_TO_CLASS_PATH` in `tfx.orchestration.kubeflow.v2.compiler_utils`, mapping `system.*` and `tfx.*` schema titles to TFX Python class paths.

**Why**: KFP v2 backend remaps TFX artifact types to generic `system.*` titles (e.g. `system.Model`, `system.Dataset`). When TFX tries to reconstruct artifact objects, it attempts to import these titles as Python module paths (e.g. `import system.Model`), which fails with `ImportError`. The mapping tells TFX which actual class to instantiate.

**Confidence: 8/10**

This is a clean, declarative fix. The only concern is maintaining the mapping if new TFX artifact types are added, but for the penguin quickstart pipeline this covers all necessary types. An alternative would be to patch the KFP v2 backend to preserve original TFX type names, but that's a much larger upstream change.

---

## Fix 2: Null Schema Handling (`kubeflow_v2_entrypoint_utils`)

**What**: Replace `yaml.safe_load(type_schema.instance_schema)` with `yaml.safe_load(type_schema.instance_schema) or {}` to handle `None` return values.

**Why**: KFP v2 driver sometimes passes empty `instance_schema` strings, causing `yaml.safe_load` to return `None`. Downstream code then fails when trying to access dictionary methods on `None`.

**Confidence: 9/10**

This is a defensive fix with no downside — converting `None` to `{}` is the correct fallback. The only alternative would be fixing the KFP v2 driver to never emit empty schemas, which would be the ideal upstream fix.

---

## Fix 3: Artifact Class Inference from Metadata (`kubeflow_v2_entrypoint_utils`)

**What**: Patch `_parse_raw_artifact` to inspect artifact metadata keys (e.g. `split_names`, `span`) and infer the correct TFX artifact class when the schema title is a generic `system.Dataset`.

**Why**: KFP v2 maps multiple TFX types to a single system type (e.g. both `tfx.Examples` and `tfx.Dataset` become `system.Dataset`). Without this, TFX creates generic `Dataset` objects instead of `Examples`, losing type-specific properties.

**Confidence: 5/10**

This is a heuristic based on checking metadata key signatures. It works for known TFX types but is fragile — new artifact types or changed metadata schemas could break it.

**Alternative proposal (7/10)**: Encode the original TFX type name in a custom metadata field during compilation (e.g. `_tfx_type: "tfx.Examples"`), then read it back during execution. This would be deterministic rather than heuristic, but requires changes to the TFX compiler as well.

---

## Fix 4: Input Artifact Type Enrichment (`kubeflow_v2_run_executor`)

**What**: Before `parse_raw_artifact_dict` is called, copy the artifact type from the component's `inputs_spec` to each input artifact that has an empty or missing `instanceSchema`.

**Why**: The KFP v2 driver does not populate `instanceSchema` on input artifacts resolved from upstream MLMD outputs (there's a TODO in the KFP driver's `metadata/client.go`). This causes TFX to create generic `Artifact` objects that lack type-specific properties like `split_names`, leading to `AttributeError: Artifact has no property 'split_names'` when components try to access them.

**Confidence: 8/10**

This leverages the `inputs_spec` which already contains the correct type information — it's authoritative data, not a heuristic. The proper fix would be in the KFP v2 driver itself (the TODO referenced in the Go code), but that's an upstream change. This patch is a reliable workaround.

---

## Fix 5: Remove `Format-Serving` Subdirectory (`path_utils`)

**What**: Patch `tfx.utils.path_utils.serving_model_dir` to return `output_uri` directly instead of `os.path.join(output_uri, SERVING_MODEL_DIR)`.

**Why**: TFX by default saves trained models under `model_uri/Format-Serving/`. This creates a nested directory structure in GCS that triggers a bug in the KFP v2 launcher's artifact download code (see Fix 6). Flattening the model to save directly under `model_uri/` reduces nesting depth.

**Confidence: 4/10**

This changes TFX's expected model layout, which could break other components or tools that expect the `Format-Serving` subdirectory convention (e.g. `Evaluator`, custom model serving logic, TFX model validation). It works for the quickstart pipeline but is risky for more complex pipelines.

**Alternative proposal (7/10)**: Instead of changing where the model is saved, fix the root cause in the KFP v2 launcher (Fix 6 below handles the symptom). A proper fix would be to patch the KFP v2 launcher's Go code to handle GCS directory markers correctly — either by sorting blobs to process directories before files, or by detecting and skipping zero-byte marker blobs. This would be an upstream KFP contribution but would fix the problem for all artifact types, not just models.

---

## Fix 6: GCS Directory Marker Cleanup (`trainer.py`)

**What**: After saving the model, iterate over all GCS blobs under the model URI and delete any zero-byte objects whose names end with `/` (directory markers).

**Why**: The KFP v2 launcher downloads all GCS blobs under an artifact URI to a local filesystem. GCS "directory markers" (zero-byte blobs ending with `/`) get downloaded as regular files, then when the launcher tries to create subdirectories at the same path, it fails with `mkdir: not a directory`. Deleting these markers after the Trainer saves ensures the Pusher can download the model artifact cleanly.

**Confidence: 5/10**

This works but has limitations:
- Only cleans up markers created by the Trainer — other components with directory artifacts could hit the same issue
- Requires GCS client library access and appropriate permissions
- Race condition risk if something recreates markers between cleanup and download (unlikely but possible)
- Couples the trainer user code to infrastructure concerns

**Alternative proposal (9/10)**: Fix the KFP v2 launcher's `downloadFile()` function in Go to skip zero-byte blobs ending with `/`, or to create directories (not files) for such blobs. This is a ~10-line change in the launcher code and would fix the problem globally for all components and artifact types. This should be contributed upstream to KFP.

---

## Summary Table

| Fix | Target | Confidence | Risk |
|-----|--------|:----------:|------|
| 1. Schema title mapping | `compiler_utils` | 8/10 | Low — declarative mapping |
| 2. Null schema handling | `entrypoint_utils` | 9/10 | Minimal — defensive null check |
| 3. Artifact class inference | `entrypoint_utils` | 5/10 | Medium — heuristic-based |
| 4. Input artifact enrichment | `run_executor` | 8/10 | Low — uses authoritative spec data |
| 5. Remove Format-Serving | `path_utils` | 4/10 | High — changes TFX convention |
| 6. GCS marker cleanup | `trainer.py` | 5/10 | Medium — symptom-level fix |

**Overall assessment**: Fixes 1, 2, and 4 are solid and appropriate for production use. Fixes 3, 5, and 6 are workarounds that should ideally be replaced by upstream fixes in the KFP v2 driver and launcher. The highest-impact upstream contribution would be fixing the KFP v2 launcher's GCS download logic, which would eliminate the need for both Fix 5 and Fix 6.
