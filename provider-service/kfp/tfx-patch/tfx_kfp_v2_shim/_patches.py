"""Monkey-patch functions applied to TFX modules after import.

Each ``patch_*`` function receives a loaded module and modifies it in-place.
These are the runtime equivalents of the source-level patches that were
previously applied by ``patch_tfx_kfp_v2.py``.

Patches
-------
1. ``patch_compiler_utils``  — add system.* → TFX class mappings
2. ``patch_entrypoint_utils`` — null-safe schema + type inference + enrichment
5. ``patch_path_utils``       — flatten model serving directory
7. ``patch_run_executor``     — clean up zero-byte directory markers after execution
"""

from __future__ import annotations

import inspect
import logging
from types import ModuleType

log = logging.getLogger(__name__)


# ── Generic Artifact stand-in for system.Artifact ─────────────────────────
# The base ``artifact.Artifact`` class cannot be instantiated without
# ``mlmd_artifact_type`` (TYPE_NAME is not set).  TFX 1.14 had
# ``simple_artifacts.File`` for this role; TFX 1.15 removed it.
# We provide a minimal concrete subclass so ``artifact_cls()`` succeeds
# inside TFX's ``_parse_raw_artifact``.
_GenericArtifact = None  # populated lazily by _get_generic_artifact_class()


def _get_generic_artifact_class():
    """Return (and cache) a concrete Artifact subclass with TYPE_NAME set."""
    global _GenericArtifact  # noqa: PLW0603
    if _GenericArtifact is None:
        from tfx.types import artifact as _artifact_mod
        _GenericArtifact = type(
            "GenericArtifact",
            (_artifact_mod.Artifact,),
            {"TYPE_NAME": "GenericArtifact"},
        )
        # Make it importable via name_utils.get_full_name → import_class_by_path
        _GenericArtifact.__module__ = __name__
        _GenericArtifact.__qualname__ = "GenericArtifact"
        # Register on this module so import_class_by_path can find it
        import sys
        sys.modules[__name__].GenericArtifact = _GenericArtifact
    return _GenericArtifact


# ── Type inference rules (patch 3) ─────────────────────────────────────────
# Disambiguate TFX types that share the same system.* title.
# Keys are the (possibly wrong) title; values are (metadata_keys, correct_title).
TYPE_INFERENCE_RULES: dict[str, list[tuple[frozenset, str]]] = {
    "system.Dataset": [
        (frozenset({"split_names", "span", "version"}), "tfx.Examples"),
        (frozenset({"split_names", "span"}), "tfx.Examples"),
        (frozenset({"split_names"}), "tfx.Examples"),
    ],
    "system.Model": [
        (frozenset({"pushed_destination", "pushed"}), "tfx.PushedModel"),
        (frozenset({"pushed_destination"}), "tfx.PushedModel"),
        (frozenset({"pushed"}), "tfx.PushedModel"),
    ],
}
# Also match tfx.* variants that may appear via instance_schema titles.
TYPE_INFERENCE_RULES["tfx.Dataset"] = TYPE_INFERENCE_RULES["system.Dataset"]
TYPE_INFERENCE_RULES["tfx.Model"] = TYPE_INFERENCE_RULES["system.Model"]


# ── Patch 1: Schema title mapping (compiler_utils) ────────────────────────

def patch_compiler_utils(mod: ModuleType) -> None:
    """Add system.* artifact type entries to TITLE_TO_CLASS_PATH.

    KFP v2 normalises TFX artifact types to system.* titles (e.g. system.Model).
    TFX looks these up in TITLE_TO_CLASS_PATH but only has tfx.* entries.
    """
    from tfx.types import standard_artifacts
    from tfx.utils import name_utils

    # TFX 1.14 has simple_artifacts; TFX 1.15+ moved them to system_artifacts
    try:
        from tfx.types import simple_artifacts
        file_cls = simple_artifacts.File
        dataset_cls = simple_artifacts.Dataset
        metrics_cls = simple_artifacts.Metrics
        statistics_cls = simple_artifacts.Statistics
    except ImportError:
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
    mod.TITLE_TO_CLASS_PATH.update(
        {title: name_utils.get_full_name(cls) for title, cls in system_to_tfx.items()}
    )
    added = {title: name_utils.get_full_name(cls) for title, cls in system_to_tfx.items()}
    log.info("Patch 1: added %d system.* title mappings: %s", len(added), added)
    log.info("Patch 1: TITLE_TO_CLASS_PATH id=%d, keys=%s", id(mod.TITLE_TO_CLASS_PATH), list(mod.TITLE_TO_CLASS_PATH.keys()))


# ── Patches 2, 3, 4: entrypoint_utils ─────────────────────────────────────

def patch_entrypoint_utils(mod: ModuleType) -> None:
    """Apply patches 2 (null schema), 3 (type inference), 4 (enrichment), 7 (GCS markers)."""
    _patch_parse_raw_artifact(mod)
    _patch_parse_raw_artifact_dict(mod)
    _patch_translate_executor_output(mod)


def _patch_parse_raw_artifact(mod: ModuleType) -> None:
    """Wrap ``_parse_raw_artifact`` with null-schema handling + type inference.

    Patch 2: ``yaml.safe_load("")`` returns None → guard with ``or {}``.
    Patch 3: Infer correct TFX type from metadata keys before delegating.
    """
    import yaml

    original = mod._parse_raw_artifact

    def _patched(artifact_pb, name_from_id):
        type_schema = artifact_pb.type
        kind = type_schema.WhichOneof("kind")

        # Resolve current title (with null-safe yaml load — patch 2).
        title = None
        if kind == "schema_title":
            title = type_schema.schema_title
        elif kind == "instance_schema":
            title = (yaml.safe_load(type_schema.instance_schema) or {}).get("title")

        # Infer correct TFX type from metadata keys (patch 3).
        rules = TYPE_INFERENCE_RULES.get(title)
        if rules:
            all_keys = _collect_artifact_keys(artifact_pb)
            for signature, tfx_title in rules:
                if signature <= all_keys:
                    log.debug("Inferred %s from keys %s", tfx_title, all_keys)
                    if kind == "schema_title":
                        artifact_pb.type.schema_title = tfx_title
                    break

        # Guard null instance_schema so the original won't crash (patch 2).
        if kind == "instance_schema" and not type_schema.instance_schema:
            artifact_pb.type.instance_schema = "title: ''"

        # Debug: log the type schema being processed
        log.info(
            "Patch 2+3: processing artifact kind=%s title=%s schema=%r",
            kind, title,
            getattr(type_schema, 'instance_schema', None)
            if kind == 'instance_schema'
            else getattr(type_schema, 'schema_title', None),
        )

        try:
            return original(artifact_pb, name_from_id)
        except (ValueError, ImportError) as exc:
            log.error(
                "Patch 2+3: original _parse_raw_artifact failed for "
                "kind=%s title=%s schema=%r: %s",
                kind, title,
                getattr(type_schema, 'instance_schema', None)
                if kind == 'instance_schema'
                else getattr(type_schema, 'schema_title', None),
                exc,
            )
            raise

    mod._parse_raw_artifact = _patched
    log.info("Patch 2+3: wrapped _parse_raw_artifact")


def _patch_parse_raw_artifact_dict(mod: ModuleType) -> None:
    """Wrap ``parse_raw_artifact_dict`` to enrich input artifacts (patch 4).

    KFP v2 driver resolves input artifacts but may not populate their type.
    The component's ``inputs_spec`` has the authoritative type info.  This
    wrapper inspects the caller's frame for ``inputs_spec`` and copies
    artifact type schemas into artifacts that lack them.
    """
    original = mod.parse_raw_artifact_dict

    def _patched(artifact_dict, name_from_id):
        # Look for inputs_spec in the caller's local variables.
        frame = inspect.currentframe()
        try:
            caller_locals = frame.f_back.f_locals if frame.f_back else {}
        finally:
            del frame  # avoid reference cycles

        inputs_spec = caller_locals.get("inputs_spec")
        if inputs_spec:
            _enrich_input_artifacts(artifact_dict, inputs_spec)

        return original(artifact_dict, name_from_id)

    mod.parse_raw_artifact_dict = _patched
    log.info("Patch 4: wrapped parse_raw_artifact_dict")


def _patch_translate_executor_output(mod: ModuleType) -> None:
    """Wrap ``translate_executor_output`` to clean GCS markers (patch 7).

    After each executor finishes, ``translate_executor_output`` is called with
    the output artifacts.  We wrap it to delete zero-byte GCS directory
    markers from each output artifact URI before the launcher uploads them.
    """
    original = mod.translate_executor_output

    def _patched(output_dict, name_from_id):
        # Clean up GCS markers from output artifact URIs.
        for _key, artifacts in output_dict.items():
            for art in artifacts:
                uri = art.uri
                if uri and uri.startswith("gs://"):
                    _delete_gcs_directory_markers(uri)

        return original(output_dict, name_from_id)

    mod.translate_executor_output = _patched
    log.info("Patch 7: wrapped translate_executor_output with GCS marker cleanup")


# ── Patch 6: Re-export KubeflowV2DagRunner in experimental ────────────────

def patch_experimental(mod: ModuleType) -> None:
    """Re-export KubeflowV2DagRunner in tfx.v1.orchestration.experimental.

    With kfp>=2, TFX removes KubeflowV2DagRunner from the experimental
    namespace.  The kfp-operator compiler imports it from there, so we
    re-inject it from the v2 runner module.
    """
    if hasattr(mod, "KubeflowV2DagRunner"):
        log.info("Patch 6: KubeflowV2DagRunner already present, skipping")
        return

    try:
        from tfx.orchestration.kubeflow.v2 import kubeflow_v2_dag_runner
        mod.KubeflowV2DagRunner = kubeflow_v2_dag_runner.KubeflowV2DagRunner
        mod.KubeflowV2DagRunnerConfig = kubeflow_v2_dag_runner.KubeflowV2DagRunnerConfig
        log.info("Patch 6: re-exported KubeflowV2DagRunner into experimental")
    except (ImportError, AttributeError) as exc:
        log.warning("Patch 6: could not import KubeflowV2DagRunner: %s", exc)


# ── Patch 5: Flatten model directory (path_utils) ─────────────────────────

def patch_path_utils(mod: ModuleType) -> None:
    """Replace ``serving_model_dir`` to skip the Format-Serving subdirectory.

    KFP v2 launcher's DownloadBlob chokes on TFX's nested directory model
    layout because GCS zero-byte directory markers get downloaded as files,
    blocking os.MkdirAll.  Returning ``output_uri`` directly avoids this.

    WARNING: This breaks Vertex AI compatibility.
    """
    def _flat_serving_model_dir(output_uri: str, is_old_artifact: bool = False) -> str:
        return output_uri

    mod.serving_model_dir = _flat_serving_model_dir
    log.info("Patch 5: flattened serving_model_dir")


# ── Patch 7: zero-byte directory marker cleanup (run_executor) ─────────────────

def patch_run_executor(mod: ModuleType) -> None:
    """Wrap ``_run_executor`` to delete zero-byte GCS directory markers.

    TensorFlow's legacy GCS filesystem (TF ≤ 2.16) creates zero-byte
    placeholder blobs as directory markers (e.g. ``gs://…/model/``).
    The KFP v2 launcher downloads these as regular files, which then
    block ``os.MkdirAll`` for child paths in downstream components.

    This patch runs after each component executor completes and deletes
    any zero-byte blobs whose names end with ``/`` under each output
    artifact URI.
    """
    original = mod._run_executor

    def _patched_run_executor(args, beam_args):
        original(args, beam_args)
        _cleanup_output_markers(args)

    mod._run_executor = _patched_run_executor
    log.info("Patch 7: wrapped _run_executor with GCS marker cleanup")


def _cleanup_output_markers(args) -> None:
    """Parse output artifact URIs from executor args and remove markers."""
    try:
        from google.protobuf import json_format
        from kfp.pipeline_spec import pipeline_spec_pb2

        executor_input = pipeline_spec_pb2.ExecutorInput()
        json_format.Parse(
            args.json_serialized_invocation_args,
            executor_input,
            ignore_unknown_fields=True,
        )
        for _key, artifact_list in executor_input.outputs.artifacts.items():
            for artifact in artifact_list.artifacts:
                uri = artifact.uri
                if uri and uri.startswith("gs://"):
                    _delete_gcs_directory_markers(uri)
    except Exception as exc:
        log.warning("Patch 7: marker cleanup failed: %s", exc)


def _delete_gcs_directory_markers(uri: str) -> None:
    """Delete zero-byte blobs ending with ``/`` under *uri*."""
    try:
        from google.cloud import storage as gcs_storage

        path = uri[5:]  # strip "gs://"
        bucket_name, _, prefix = path.partition("/")
        if prefix and not prefix.endswith("/"):
            prefix += "/"

        client = gcs_storage.Client()
        bucket = client.bucket(bucket_name)

        deleted = 0
        for blob in bucket.list_blobs(prefix=prefix):
            if blob.size == 0 and blob.name.endswith("/"):
                blob.delete()
                deleted += 1

        if deleted:
            log.info(
                "Patch 7: deleted %d GCS directory marker(s) under %s",
                deleted, uri,
            )
    except Exception as exc:
        log.warning(
            "Patch 7: could not clean markers under %s: %s", uri, exc,
        )


# ── Helpers ────────────────────────────────────────────────────────────────

def _collect_artifact_keys(artifact_pb) -> set:
    """Collect all metadata + property keys from a RuntimeArtifact proto."""
    keys: set = set()
    if hasattr(artifact_pb, "metadata") and artifact_pb.metadata:
        fields = getattr(artifact_pb.metadata, "fields", None)
        if fields:
            keys.update(fields.keys())
    if hasattr(artifact_pb, "properties") and artifact_pb.properties:
        keys.update(artifact_pb.properties.keys())
    return keys


def _enrich_input_artifacts(artifact_dict, inputs_spec) -> None:
    """Copy artifact type schemas from inputs_spec into untyped artifacts."""
    for key, artifact_spec in inputs_spec.artifacts.items():
        if key not in artifact_dict:
            continue
        for art in artifact_dict[key].artifacts:
            kind = art.type.WhichOneof("kind")
            has_empty_schema = (
                kind == "instance_schema" and not art.type.instance_schema
            )
            if not kind or has_empty_schema:
                art.type.CopyFrom(artifact_spec.artifact_type)
                log.debug("Enriched artifact %s with type from inputs_spec", key)
