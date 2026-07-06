"""Monkey-patches applied to TFX modules via the import hook in ``_hook.py``."""

from __future__ import annotations

import functools
import inspect
import logging
import os
import sys
from types import ModuleType

log = logging.getLogger(__name__)


# TFX 1.15 removed simple_artifacts.File but artifact.Artifact can't be
# instantiated without TYPE_NAME. This lazy subclass fills the gap.

@functools.lru_cache(maxsize=1)
def _get_generic_artifact_class():
    """Return a concrete Artifact subclass usable as a system.Artifact stand-in."""
    from tfx.types import artifact as _artifact_mod

    cls = type(
        "GenericArtifact",
        (_artifact_mod.Artifact,),
        {"TYPE_NAME": "GenericArtifact"},
    )
    cls.__module__ = __name__
    cls.__qualname__ = "GenericArtifact"
    # Register on this module so name_utils.get_full_name() can resolve it and
    # import_utils.import_class_by_path() can re-import it at executor runtime;
    # otherwise the synthetic class is not reachable as an attribute of its
    # declared module and TFX rejects it as "not importable".
    setattr(sys.modules[__name__], "GenericArtifact", cls)
    return cls


# KFP v2 collapses multiple TFX types into the same system.* title. These
# rules disambiguate by checking which metadata keys are present. Matching is
# by subset (see _patch_parse_raw_artifact), so only the minimal distinguishing
# key set is listed — a larger key set on the artifact still matches.
TYPE_INFERENCE_RULES: dict[str, list[tuple[frozenset, str]]] = {
    # Examples carry split_names (alongside span/version); other Datasets do not.
    "system.Dataset": [
        (frozenset({"split_names"}), "tfx.Examples"),
    ],
    # PushedModel carries pushed and/or pushed_destination; plain Models do not.
    "system.Model": [
        (frozenset({"pushed_destination"}), "tfx.PushedModel"),
        (frozenset({"pushed"}), "tfx.PushedModel"),
    ],
}
TYPE_INFERENCE_RULES["tfx.Dataset"] = TYPE_INFERENCE_RULES["system.Dataset"]
TYPE_INFERENCE_RULES["tfx.Model"] = TYPE_INFERENCE_RULES["system.Model"]


# ── Patch 1 ──────────────────────────────────────────────────────────────
# Issue:   KFP v2 normalises artifact types to system.* titles (e.g. system.Model).
#          TFX's TITLE_TO_CLASS_PATH only has tfx.* entries, so lookup fails with
#          ImportError ("import system.Model").
# Solution: Add system.* → TFX class mappings to the lookup table.

def patch_compiler_utils(mod: ModuleType) -> None:
    """Add system.* entries to TITLE_TO_CLASS_PATH."""
    from tfx.types import standard_artifacts
    from tfx.utils import name_utils

    try:
        from tfx.types import simple_artifacts
        file_cls = simple_artifacts.File
        dataset_cls = simple_artifacts.Dataset
        metrics_cls = simple_artifacts.Metrics
        statistics_cls = simple_artifacts.Statistics
    except ImportError:  # TFX 1.15+
        from tfx.types import system_artifacts
        file_cls = _get_generic_artifact_class()
        dataset_cls = system_artifacts.Dataset
        metrics_cls = system_artifacts.Metrics
        statistics_cls = system_artifacts.Statistics

    system_to_tfx = {
        "system.Artifact": file_cls,
        "system.Model": standard_artifacts.Model,
        "system.Dataset": dataset_cls,
        "system.Metrics": metrics_cls,
        "system.Statistics": statistics_cls,
    }
    added = {title: name_utils.get_full_name(cls) for title, cls in system_to_tfx.items()}
    mod.TITLE_TO_CLASS_PATH.update(added)


# ── Patches 2, 3, 4, 7 ──────────────────────────────────────────────────

def patch_entrypoint_utils(mod: ModuleType) -> None:
    """Apply all entrypoint_utils patches."""
    _patch_parse_raw_artifact(mod)
    _patch_parse_raw_artifact_dict(mod)
    _patch_translate_executor_output(mod)


# ── Patch 2 ──────────────────────────────────────────────────────────────
# Issue:   KFP v2 driver passes artifacts with empty instance_schema.
#          yaml.safe_load("") returns None, causing TypeError on .get("title").
# Solution: Guard with ``or {}`` so None becomes an empty dict.
#
# ── Patch 3 ──────────────────────────────────────────────────────────────
# Issue:   KFP v2 collapses distinct TFX types to the same system.* title
#          (e.g. Examples and Dataset both become system.Dataset).
# Solution: Inspect artifact metadata keys to infer the correct TFX type.

def _patch_parse_raw_artifact(mod: ModuleType) -> None:
    """Wrap _parse_raw_artifact: guard null schema and infer types from keys."""
    import yaml

    original = mod._parse_raw_artifact

    def _patched(artifact_pb, name_from_id):
        type_schema = artifact_pb.type
        kind = type_schema.WhichOneof("kind")

        title = None
        if kind == "schema_title":
            title = type_schema.schema_title
        elif kind == "instance_schema":
            title = (yaml.safe_load(type_schema.instance_schema) or {}).get("title")

        rules = TYPE_INFERENCE_RULES.get(title)
        if rules:
            all_keys = _collect_artifact_keys(artifact_pb)
            for signature, tfx_title in rules:
                if signature <= all_keys:
                    log.debug("Inferred %s from keys %s", tfx_title, all_keys)
                    if kind == "schema_title":
                        artifact_pb.type.schema_title = tfx_title
                    break

        if kind == "instance_schema" and not type_schema.instance_schema:
            artifact_pb.type.instance_schema = "title: ''"

        try:
            return original(artifact_pb, name_from_id)
        except (ValueError, ImportError) as exc:
            log.error("_parse_raw_artifact failed: kind=%s title=%s: %s", kind, title, exc)
            raise

    mod._parse_raw_artifact = _patched


# ── Patch 4 ──────────────────────────────────────────────────────────────
# Issue:   KFP v2 driver resolves input artifacts from MLMD but does not
#          populate their type schema. TFX creates generic Artifact objects
#          that lack type-specific properties (e.g. split_names).
# Solution: Read inputs_spec from the caller's frame and copy artifact type
#           schemas into untyped artifacts before TFX parses them.

def _patch_parse_raw_artifact_dict(mod: ModuleType) -> None:
    """Wrap parse_raw_artifact_dict: enrich untyped artifacts from inputs_spec."""
    original = mod.parse_raw_artifact_dict

    def _patched(artifact_dict, name_from_id):
        frame = inspect.currentframe()
        try:
            caller_locals = frame.f_back.f_locals if frame.f_back else {}
        finally:
            del frame

        inputs_spec = caller_locals.get("inputs_spec")
        if inputs_spec:
            _enrich_input_artifacts(artifact_dict, inputs_spec)

        return original(artifact_dict, name_from_id)

    mod.parse_raw_artifact_dict = _patched


# ── Patch 7 ──────────────────────────────────────────────────────────────
# Issue:   TF's legacy GCS filesystem (≤2.16) creates zero-byte directory
#          marker blobs. The KFP launcher downloads them as files, blocking
#          os.MkdirAll for child paths in downstream components.
# Solution: After each executor run, list output artifact URIs and delete
#           any zero-byte blobs whose name ends with "/".

def _patch_translate_executor_output(mod: ModuleType) -> None:
    """Wrap translate_executor_output: delete GCS directory markers from outputs."""
    original = mod.translate_executor_output

    @functools.lru_cache(maxsize=1)
    def _get_client():
        from google.cloud import storage as gcs_storage
        return gcs_storage.Client()

    def _patched(output_dict, name_from_id):
        for _key, artifacts in output_dict.items():
            for art in artifacts:
                if art.uri and art.uri.startswith("gs://"):
                    _delete_gcs_directory_markers(art.uri, _get_client())
        return original(output_dict, name_from_id)

    mod.translate_executor_output = _patched


# ── Patch 5 ──────────────────────────────────────────────────────────────
# Issue:   TFX saves models under <uri>/Format-Serving/. The KFP launcher's
#          DownloadBlob treats GCS directory markers as files, which blocks
#          os.MkdirAll for the nested child paths.
# Solution: Replace serving_model_dir to return output_uri directly,
#           removing the nested subdirectory. Breaks Vertex AI.

def patch_path_utils(mod: ModuleType) -> None:
    """Skip Format-Serving/ subdirectory. Breaks Vertex AI."""
    def _flat(output_uri: str, is_old_artifact: bool = False) -> str:
        return output_uri

    mod.serving_model_dir = _flat

# ── Patch 6 ──────────────────────────────────────────────────────────────
# Issue:   With kfp>=2, TFX removes KubeflowV2DagRunner from
#          tfx.v1.orchestration.experimental. The kfp-operator compiler
#          imports it from that path.
# Solution: Re-export it from tfx.orchestration.kubeflow.v2.

def patch_experimental(mod: ModuleType) -> None:
    """Re-inject KubeflowV2DagRunner removed by kfp>=2."""
    if hasattr(mod, "KubeflowV2DagRunner"):
        return

    try:
        from tfx.orchestration.kubeflow.v2 import kubeflow_v2_dag_runner
        mod.KubeflowV2DagRunner = kubeflow_v2_dag_runner.KubeflowV2DagRunner
        mod.KubeflowV2DagRunnerConfig = kubeflow_v2_dag_runner.KubeflowV2DagRunnerConfig
        log.info("Re-exported KubeflowV2DagRunner")
    except (ImportError, AttributeError) as exc:
        log.warning("KubeflowV2DagRunner unavailable: %s", exc)


# ── Patch 8 ──────────────────────────────────────────────────────────────
# Issue:   On container exit after a TFX component runs, the interpreter
#          shutdown sequence triggers C++ destructors in protobuf/MLMD that
#          crash with "pure virtual method called" (seen on TFX 1.17.3 / S3).
# Solution: Wrap _run_executor to call os._exit(0) once it returns, bypassing
#           the shutdown that provokes the crash. The metadata write is the
#           final statement of _run_executor, so exiting after it returns is
#           behaviourally identical to exiting immediately after the write.

def patch_run_executor(mod: ModuleType) -> None:
    """Wrap _run_executor to force-exit(0) after it completes.

    NOTE: the executor container runs this module via
    ``python -m ...kubeflow_v2_run_executor`` (i.e. as ``__main__``), which the
    import hook cannot patch — runpy loads it through the loader's ``get_code``,
    bypassing ``exec_module``. The effective force-exit for that entry path is a
    source edit applied at install time (see install_shim.patch_run_executor_source).
    This runtime wrapper only covers the (rare) case where the module is imported
    and ``_run_executor`` is called directly.
    """
    original = mod._run_executor

    @functools.wraps(original)
    def _patched(*args, **kwargs):
        original(*args, **kwargs)
        log.info("Executor complete; forcing os._exit(0) to avoid shutdown crash")
        os._exit(0)

    mod._run_executor = _patched


# ── Helpers ──────────────────────────────────────────────────────────────

def _delete_gcs_directory_markers(uri: str, client) -> None:
    """Delete zero-byte blobs ending with ``/`` under *uri*."""
    try:
        path = uri[5:]
        bucket_name, _, prefix = path.partition("/")
        if prefix and not prefix.endswith("/"):
            prefix += "/"

        bucket = client.bucket(bucket_name)
        markers = [
            blob for blob in bucket.list_blobs(prefix=prefix)
            if blob.size == 0 and blob.name.endswith("/")
        ]

        if markers:
            with client.batch():
                for blob in markers:
                    blob.delete()
            log.info("Deleted %d GCS directory marker(s) under %s", len(markers), uri)
    except Exception as exc:
        log.warning("GCS marker cleanup failed for %s: %s", uri, exc)


def _collect_artifact_keys(artifact_pb) -> set:
    """Return all metadata + property keys from a RuntimeArtifact."""
    keys: set = set()
    metadata = getattr(artifact_pb, "metadata", None)
    if metadata and getattr(metadata, "fields", None):
        keys.update(metadata.fields.keys())

    properties = getattr(artifact_pb, "properties", None)
    if properties:
        keys.update(properties.keys())
    return keys


def _enrich_input_artifacts(artifact_dict, inputs_spec) -> None:
    """Copy type schemas from inputs_spec into artifacts that lack one."""
    for key, artifact_spec in inputs_spec.artifacts.items():
        if key not in artifact_dict:
            continue
        for art in artifact_dict[key].artifacts:
            kind = art.type.WhichOneof("kind")
            if not kind or (kind == "instance_schema" and not art.type.instance_schema):
                art.type.CopyFrom(artifact_spec.artifact_type)
                log.debug("Enriched artifact %s from inputs_spec", key)
