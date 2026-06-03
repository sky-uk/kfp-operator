"""Patch TFX (1.15–1.17) for compatibility with OSS Kubeflow Pipelines v2.

What: KFP v2's driver/launcher architecture is incompatible with TFX's runtime
      expectations. TFX's KFP v2 adapter was built for Vertex AI Pipelines,
      whose driver behaves differently. This script applies targeted patches.

Why:  Without these patches, TFX pipelines on OSS KFP v2 fail with:
        - ImportError resolving system.* artifact types
        - AttributeError on artifact properties (e.g. split_names)
        - mkdir failures downloading nested directory artifacts from GCS

How:  Each patch modifies a TFX source file in-place. Patches are idempotent
      and assert the exact code they expect before modifying anything.

Usage:
    python patch_tfx_kfp_v2.py                  # apply all patches
    python patch_tfx_kfp_v2.py --check           # dry-run verification
    python patch_tfx_kfp_v2.py --vertex-compatible  # skip patch 5
"""

from __future__ import annotations

import argparse
import importlib
import logging
import sys
from dataclasses import dataclass
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
log = logging.getLogger(__name__)

PATCH_MARKER = "# [patch:kfp-v2-compat]"

# Target modules
_COMPILER_UTILS = "tfx.orchestration.kubeflow.v2.compiler_utils"
_ENTRYPOINT_UTILS = (
    "tfx.orchestration.kubeflow.v2.container.kubeflow_v2_entrypoint_utils"
)
_RUN_EXECUTOR = (
    "tfx.orchestration.kubeflow.v2.container.kubeflow_v2_run_executor"
)
_PATH_UTILS = "tfx.utils.path_utils"


# ---------------------------------------------------------------------------
# Patch data model
# ---------------------------------------------------------------------------

@dataclass(frozen=True)
class Patch:
    """A single source-level patch to a TFX module.

    Each patch finds `anchor` in the target module and either replaces it with
    `replacement` or appends `replacement` to the end of the file.

    Attributes:
        name:             Human-readable label for logging.
        module:           Dotted Python module path to patch.
        anchor:           Exact string that must exist in the source.
        replacement:      Code to replace/append.
        append:           If True, append `replacement` instead of replacing `anchor`.
        idempotency_mark: String whose presence means "already patched".
    """

    name: str
    module: str
    anchor: str
    replacement: str
    append: bool = False
    idempotency_mark: str = PATCH_MARKER

    def apply(self, check_only: bool = False) -> None:
        filepath = _resolve_module_path(self.module)
        source = filepath.read_text()

        if self.idempotency_mark in source:
            log.info("%s: already applied — %s", self.name, filepath)
            return

        if self.anchor not in source:
            raise AssertionError(
                f"Expected code pattern not found in {filepath}.\n"
                f"This patch is designed for TFX 1.15–1.17.\n"
                f"Expected:\n{self.anchor}"
            )

        if check_only:
            log.info("%s: ready to apply — %s", self.name, filepath)
            return

        if self.append:
            filepath.write_text(source + self.replacement)
        else:
            filepath.write_text(source.replace(self.anchor, self.replacement, 1))

        log.info("%s: applied — %s", self.name, filepath)


def _resolve_module_path(module_path: str) -> Path:
    """Return the on-disk source file for an installed Python module."""
    mod = importlib.import_module(module_path)
    path = Path(mod.__file__)
    if not path.exists():
        raise FileNotFoundError(f"Module source not found: {path}")
    return path


# ---------------------------------------------------------------------------
# Patch definitions
# ---------------------------------------------------------------------------

# Patch 1 — Schema title mapping (compiler_utils.py)
#
# KFP v2 normalises artifact types into system.* titles (e.g. system.Model).
# TFX looks these up in TITLE_TO_CLASS_PATH to find the Python class, but the
# table only has tfx.* entries. system.* titles fall through to
# import_class_by_path which tries "import system.Model" and fails.
# Fix: add system.* → TFX class mappings to the lookup table.

_PATCH_1 = Patch(
    name="1. Schema title mapping",
    module=_COMPILER_UTILS,
    anchor=(
        "TITLE_TO_CLASS_PATH = {\n"
        "    f'tfx.{klass.__qualname__}': name_utils.get_full_name(klass)\n"
        "    for klass in _SUPPORTED_STANDARD_ARTIFACT_TYPES\n"
        "}"
    ),
    replacement=(
        "TITLE_TO_CLASS_PATH = {\n"
        "    f'tfx.{klass.__qualname__}': name_utils.get_full_name(klass)\n"
        "    for klass in _SUPPORTED_STANDARD_ARTIFACT_TYPES\n"
        "}\n"
        f"\n{PATCH_MARKER}\n"
        "_SYSTEM_TO_TFX_CLASS = {\n"
        '    "system.Artifact": simple_artifacts.File,\n'
        '    "system.Model": standard_artifacts.Model,\n'
        '    "system.Dataset": simple_artifacts.Dataset,\n'
        '    "system.Metrics": simple_artifacts.Metrics,\n'
        '    "system.Statistics": simple_artifacts.Statistics,\n'
        "}\n"
        "TITLE_TO_CLASS_PATH.update({\n"
        "    title: name_utils.get_full_name(cls)\n"
        "    for title, cls in _SYSTEM_TO_TFX_CLASS.items()\n"
        "})"
    ),
)

# Patch 2 — Null instance_schema handling (entrypoint_utils.py)
#
# KFP v2 driver sometimes passes artifacts with an empty instance_schema.
# yaml.safe_load("") returns None, causing TypeError on .get("title").
# Fix: append "or {}" so None becomes an empty dict.

_PATCH_2 = Patch(
    name="2. Null schema handling",
    module=_ENTRYPOINT_UTILS,
    anchor="data = yaml.safe_load(type_schema.instance_schema)",
    replacement=(
        "data = yaml.safe_load(type_schema.instance_schema)"
        f" or {{}}  {PATCH_MARKER}"
    ),
)

# Patch 3 — Artifact class inference from metadata (entrypoint_utils.py)
#
# KFP v2 maps multiple TFX types to the same system.* title (e.g. both
# tfx.Examples and tfx.Dataset become system.Dataset). After patch 1 maps
# system.Dataset → Dataset, artifacts lack Examples-specific properties like
# split_names. Fix: inspect metadata keys to infer the correct TFX type
# before delegating to the original _parse_raw_artifact.

_PATCH_3_APPENDED_CODE = f"""

{PATCH_MARKER}
_DATASET_SIGNATURES = [
    (frozenset({{"split_names", "span", "version"}}), "tfx.Examples"),
    (frozenset({{"split_names", "span"}}), "tfx.Examples"),
    (frozenset({{"split_names"}}), "tfx.Examples"),
]

_original_parse_raw_artifact = _parse_raw_artifact


def _parse_raw_artifact(artifact_pb, name_from_id):
    type_schema = artifact_pb.type
    kind = type_schema.WhichOneof("kind")

    title = None
    if kind == "schema_title":
        title = type_schema.schema_title
    elif kind == "instance_schema":
        data = yaml.safe_load(type_schema.instance_schema) or {{}}
        title = data.get("title")

    if title in ("system.Dataset", "tfx.Dataset"):
        meta_keys = set()
        if artifact_pb.metadata and artifact_pb.metadata.fields:
            meta_keys = set(artifact_pb.metadata.fields.keys())
        prop_keys = set(artifact_pb.properties.keys()) if artifact_pb.properties else set()
        all_keys = meta_keys | prop_keys

        for signature, tfx_title in _DATASET_SIGNATURES:
            if signature <= all_keys:
                logging.info(
                    "Inferred TFX type %s from metadata keys %s", tfx_title, all_keys
                )
                if kind == "schema_title":
                    artifact_pb.type.schema_title = tfx_title
                break

    return _original_parse_raw_artifact(artifact_pb, name_from_id)
"""

_PATCH_3 = Patch(
    name="3. Artifact class inference",
    module=_ENTRYPOINT_UTILS,
    anchor=(
        "def _parse_raw_artifact(\n"
        "    artifact_pb: pipeline_pb2.RuntimeArtifact,\n"
        "    name_from_id: MutableMapping[int, str]) -> artifact.Artifact:"
    ),
    replacement=_PATCH_3_APPENDED_CODE,
    append=True,
    idempotency_mark="_original_parse_raw_artifact",
)


# Patch 4 — Input artifact type enrichment (run_executor.py)
#
# KFP v2 driver resolves input artifacts from upstream MLMD outputs but does
# not populate their instanceSchema. TFX then creates generic Artifact objects
# lacking type-specific properties. The component's inputs_spec contains the
# authoritative type information.
# Fix: copy artifact type schemas from inputs_spec into each input artifact
# before TFX parses them.

_PATCH_4 = Patch(
    name="4. Input artifact enrichment",
    module=_RUN_EXECUTOR,
    anchor=(
        "  inputs = kubeflow_v2_entrypoint_utils.parse_raw_artifact_dict(\n"
        "      inputs_dict, name_from_id\n"
        "  )"
    ),
    replacement=(
        f"  {PATCH_MARKER}\n"
        "  if inputs_spec:\n"
        "    for _key, _artifact_spec in inputs_spec.artifacts.items():\n"
        "      if _key in inputs_dict:\n"
        "        for _art in inputs_dict[_key].artifacts:\n"
        "          _kind = _art.type.WhichOneof('kind')\n"
        "          _has_empty_schema = (\n"
        "              _kind == 'instance_schema'"
        " and not _art.type.instance_schema\n"
        "          )\n"
        "          if not _kind or _has_empty_schema:\n"
        "            _art.type.CopyFrom(_artifact_spec.artifact_type)\n"
        "\n"
        "  inputs = kubeflow_v2_entrypoint_utils.parse_raw_artifact_dict(\n"
        "      inputs_dict, name_from_id\n"
        "  )"
    ),
)

# Patch 5 — Flatten model serving directory (path_utils.py)
#
# TFX saves models under <uri>/Format-Serving/. The KFP v2 launcher has a bug
# in DownloadBlob: TensorFlow's GCS client creates zero-byte directory marker
# blobs that the launcher downloads as files, blocking os.MkdirAll for child
# directories ("mkdir: not a directory").
# Fix: return output_uri directly, skipping the Format-Serving subdirectory.
#
# WARNING: This changes TFX's model layout and breaks Vertex AI compatibility.
# Use --vertex-compatible to skip this patch.

_PATCH_5 = Patch(
    name="5. Flatten model directory",
    module=_PATH_UTILS,
    anchor="return os.path.join(output_uri, path_constants.SERVING_MODEL_DIR)",
    replacement=f"return output_uri  {PATCH_MARKER}",
)


# ---------------------------------------------------------------------------
# Patch registry
# ---------------------------------------------------------------------------

# Safe for all environments including Vertex AI.
_UNIVERSAL_PATCHES = [_PATCH_1, _PATCH_2, _PATCH_3, _PATCH_4]

# Break Vertex AI compatibility — OSS KFP v2 only.
_OSS_KFP_ONLY_PATCHES = [_PATCH_5]


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Patch TFX 1.15–1.17 for KFP v2 compatibility.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Verify patches can be applied without modifying files.",
    )
    parser.add_argument(
        "--vertex-compatible",
        action="store_true",
        help="Skip patches that break Vertex AI compatibility (patch 5).",
    )
    args = parser.parse_args()

    patches = list(_UNIVERSAL_PATCHES)
    if args.vertex_compatible:
        log.info("Vertex-compatible mode: skipping OSS-only patches.")
    else:
        patches.extend(_OSS_KFP_ONLY_PATCHES)

    failures = 0
    for patch in patches:
        try:
            patch.apply(check_only=args.check)
        except (AssertionError, FileNotFoundError) as exc:
            log.error("patch %s FAILED: %s", patch.name, exc)
            failures += 1

    if failures:
        log.error("%d patch(es) failed.", failures)
        sys.exit(1)

    action = "verified" if args.check else "applied"
    log.info("All %d patches %s successfully.", len(patches), action)


if __name__ == "__main__":
    main()