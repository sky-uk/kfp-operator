"""Install the tfx_kfp_v2_shim package into the current Python environment.

This script:
  1. Copies the shim package to site-packages.
  2. Installs a .pth file that auto-activates the import hook on Python startup.

Usage (in a Dockerfile):
    COPY --from=tfx-patch /tfx_kfp_v2_shim /tmp/tfx_kfp_v2_shim
    RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim

Flags:
    --vertex-compatible   Skip patch 5 (flatten model directory)
"""

from __future__ import annotations

import argparse
import logging
import shutil
import site
import sys
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
log = logging.getLogger(__name__)

PTH_CONTENT_ALL = "import tfx_kfp_v2_shim; tfx_kfp_v2_shim.install()\n"
PTH_CONTENT_VERTEX = (
    "import tfx_kfp_v2_shim; tfx_kfp_v2_shim.install(vertex_compatible=True)\n"
)

# ── Force-exit source patch ──────────────────────────────────────────────────
# `python -m tfx...kubeflow_v2_run_executor` runs the executor module as
# __main__, which the runtime import hook cannot patch (runpy uses the loader's
# get_code, bypassing exec_module). So the force-exit is applied as a source
# edit here at install time: insert os._exit(0) right after the output metadata
# is written, bypassing the C++ destructor chain (protobuf/ml-metadata) that
# aborts with "pure virtual method called" during interpreter shutdown.
_RUN_EXECUTOR_REL = (
    "tfx/orchestration/kubeflow/v2/container/kubeflow_v2_run_executor.py"
)
_FORCE_EXIT_MARKER = "# [tfx-kfp-v2-shim:force-exit]"
_FORCE_EXIT_ANCHOR = (
    "  fileio.makedirs(os.path.dirname(metadata_uri))\n"
    "  with fileio.open(metadata_uri, 'wb') as f:\n"
    "    f.write(json_format.MessageToJson(executor_output))"
)
_FORCE_EXIT_INSERT = _FORCE_EXIT_ANCHOR + (
    "\n\n  " + _FORCE_EXIT_MARKER + "\n"
    "  # All executor work is done and output metadata is written. Exit hard to\n"
    "  # skip the C++ destructor chain that otherwise aborts with\n"
    "  # 'pure virtual method called' during Python shutdown.\n"
    "  os.write(2, b'tfx_kfp_v2_shim: force-exit(0) after metadata write\\n')\n"
    "  os._exit(0)"
)


def patch_run_executor_source(site_packages: Path) -> bool:
    """Insert os._exit(0) after the metadata write in kubeflow_v2_run_executor.

    Returns True if the patch was applied or already present, False if the
    target/anchor could not be found (e.g. an unexpected TFX version).
    """
    target = site_packages / _RUN_EXECUTOR_REL
    if not target.exists():
        log.warning("run_executor not found at %s; skipping force-exit patch", target)
        return False
    source = target.read_text()
    if _FORCE_EXIT_MARKER in source:
        log.info("force-exit source patch already present")
        return True
    if _FORCE_EXIT_ANCHOR not in source:
        log.warning(
            "force-exit anchor not found in %s; skipping (TFX version mismatch?)",
            target,
        )
        return False
    target.write_text(source.replace(_FORCE_EXIT_ANCHOR, _FORCE_EXIT_INSERT, 1))
    log.info("Applied force-exit source patch to %s", target)
    return True


def main() -> None:
    parser = argparse.ArgumentParser(description="Install tfx_kfp_v2_shim.")
    parser.add_argument(
        "--vertex-compatible",
        action="store_true",
        help="Skip patches that break Vertex AI compatibility.",
    )
    args = parser.parse_args()

    # Find the site-packages directory
    site_packages = _get_site_packages()
    target_dir = site_packages / "tfx_kfp_v2_shim"

    # Copy the shim package
    src_dir = Path(__file__).parent
    if target_dir.exists():
        shutil.rmtree(target_dir)
    shutil.copytree(src_dir, target_dir, ignore=shutil.ignore_patterns(
        "__pycache__", "*.pyc", "tests", "install_shim.py",
    ))
    log.info("Installed shim package to %s", target_dir)

    # Write the .pth file
    pth_content = PTH_CONTENT_VERTEX if args.vertex_compatible else PTH_CONTENT_ALL
    pth_file = site_packages / "tfx_kfp_v2_shim.pth"
    pth_file.write_text(pth_content)
    log.info("Installed .pth file: %s", pth_file)
    log.info("Shim will activate on every Python invocation.")

    # Apply the force-exit as a source edit (see patch_run_executor_source).
    # The runtime hook cannot patch the `python -m` executor entry module.
    patch_run_executor_source(site_packages)


def _get_site_packages() -> Path:
    """Return the primary site-packages directory."""
    # Prefer the first user/system site-packages
    paths = site.getsitepackages()
    if not paths:
        # Fallback for venvs
        paths = [site.getusersitepackages()]
    for p in paths:
        path = Path(p)
        if path.exists():
            return path
    # If nothing exists, use the first one and create it
    path = Path(paths[0])
    path.mkdir(parents=True, exist_ok=True)
    return path


if __name__ == "__main__":
    main()
