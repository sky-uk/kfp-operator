"""Patch TFX kubeflow_v2_run_executor.py to enrich input artifact types.

KFP v2 driver does not populate instanceSchema on input artifacts resolved from
upstream MLMD outputs (see GetOutputArtifactsByExecutionId TODO in metadata
client.go). This causes TFX to create generic Artifact objects that lack
type-specific properties like split_names, span, etc.

The fix enriches input artifact types from the component inputs spec (passed via
--json_serialized_inputs_spec_args) before parsing them into TFX artifacts.
"""

import re
import site
import glob
import os
import sys


def find_executor_file():
    """Find the kubeflow_v2_run_executor.py file in site-packages."""
    patterns = [
        os.path.join(
            d,
            "tfx",
            "orchestration",
            "kubeflow",
            "v2",
            "container",
            "kubeflow_v2_run_executor.py",
        )
        for d in site.getsitepackages() + [site.getusersitepackages()]
    ]
    for pattern in patterns:
        if os.path.exists(pattern):
            return pattern
    # Fallback: search common paths
    for path in glob.glob(
        "/usr/local/lib/python*/site-packages/tfx/orchestration/kubeflow/v2/container/kubeflow_v2_run_executor.py"
    ):
        return path
    return None


def patch_file(filepath):
    """Apply the artifact type enrichment patch."""
    with open(filepath, "r") as f:
        content = f.read()

    # Check if already patched
    if "enrich_input_artifact_types" in content:
        print(f"Already patched: {filepath}")
        return

    # The patch: add artifact type enrichment before parse_raw_artifact_dict
    old_code = (
        "  inputs = kubeflow_v2_entrypoint_utils.parse_raw_artifact_dict(\n"
        "      inputs_dict, name_from_id\n"
        "  )"
    )

    new_code = (
        "  # [PATCH] Enrich input artifact types from component inputs spec.\n"
        "  # KFP v2 driver does not populate instanceSchema on input artifacts\n"
        "  # resolved from upstream MLMD outputs. Use the inputs spec to fill\n"
        "  # in the correct artifact type schemas.\n"
        "  if inputs_spec:\n"
        "    for _name, _aspec in inputs_spec.artifacts.items():\n"
        "      if _name in inputs_dict:\n"
        "        for _art in inputs_dict[_name].artifacts:\n"
        "          if (not _art.type.WhichOneof('kind') or\n"
        "              (_art.type.WhichOneof('kind') == 'instance_schema'\n"
        "               and not _art.type.instance_schema)):\n"
        "            _art.type.CopyFrom(_aspec.artifact_type)\n"
        "\n"
        "  inputs = kubeflow_v2_entrypoint_utils.parse_raw_artifact_dict(\n"
        "      inputs_dict, name_from_id\n"
        "  )"
    )

    if old_code not in content:
        print(f"WARNING: Could not find target code block in {filepath}")
        print("Attempting alternative pattern...")
        # Try without leading spaces (different TFX versions may differ)
        old_code_alt = (
            "  inputs = kubeflow_v2_entrypoint_utils.parse_raw_artifact_dict(\n"
            "      inputs_dict, name_from_id)"
        )
        if old_code_alt in content:
            old_code = old_code_alt
            new_code = new_code.rstrip(")")  # Remove trailing paren to match
            new_code += ")"
        else:
            print(f"ERROR: Cannot find parse_raw_artifact_dict call in {filepath}")
            sys.exit(1)

    content = content.replace(old_code, new_code, 1)

    with open(filepath, "w") as f:
        f.write(content)

    print(f"Successfully patched: {filepath}")


def main():
    filepath = find_executor_file()
    if not filepath:
        print("ERROR: Could not find kubeflow_v2_run_executor.py")
        sys.exit(1)
    print(f"Found TFX executor at: {filepath}")
    patch_file(filepath)


if __name__ == "__main__":
    main()
