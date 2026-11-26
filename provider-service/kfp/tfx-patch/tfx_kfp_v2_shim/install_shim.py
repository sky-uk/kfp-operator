"""Install the tfx_kfp_v2_shim package into the current Python environment.

This script:
  1. Copies the shim package to site-packages.
  2. Installs a .pth file that auto-activates the import hook on Python startup.
  3. Applies the executor force-exit source patch; exits if it cannot.

Usage (in a Dockerfile), sourcing the shim via the `tfx-shim` build context:
    COPY --from=tfx-shim . /tmp/tfx_kfp_v2_shim/
    RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim
"""

from __future__ import annotations

import logging
import shutil
import site
import sys
from pathlib import Path

log = logging.getLogger(__name__)

PTH_CONTENT = "import tfx_kfp_v2_shim; tfx_kfp_v2_shim.install()\n"

# ── Force-exit source patch ──────────────────────────────────────────────────
# `python -m tfx...kubeflow_v2_run_executor` runs the executor module as
# __main__, which the runtime import hook cannot patch (runpy uses the loader's
# get_code, bypassing exec_module). So the force-exit is applied as a source
# edit here at install time: insert os._exit(0) right after the output metadata
# is written, bypassing the C++ destructor chain (protobuf/ml-metadata) that
# aborts with "pure virtual method called" during interpreter shutdown.
_RUN_EXECUTOR_PATH = "tfx/orchestration/kubeflow/v2/container/kubeflow_v2_run_executor.py"
_FORCE_EXIT_MARKER = "# [tfx-kfp-v2-shim:force-exit]"
_FORCE_EXIT_ANCHOR = (
    "  fileio.makedirs(os.path.dirname(metadata_uri))\n"
    "  with fileio.open(metadata_uri, 'wb') as f:\n"
    "    f.write(json_format.MessageToJson(executor_output))"
)
_FORCE_EXIT_INSERT = _FORCE_EXIT_ANCHOR + (
    "\n\n  "
    + _FORCE_EXIT_MARKER
    + "\n"
    + "  # All executor work is done and output metadata is written. Exit hard to\n"
    + "  # skip the C++ destructor chain that otherwise aborts with\n"
    + "  # 'pure virtual method called' during Python shutdown.\n"
    + "  os.write(2, b'tfx_kfp_v2_shim: force-exit(0) after metadata write\\n')\n"
    + "  os._exit(0)"
)


def _get_site_packages() -> Path:
    """Return a site-packages directory, creating it if needed."""
    paths = site.getsitepackages()
    if not paths:
        # Fallback for venvs
        paths = [site.getusersitepackages()]

    for p in paths:
        path = Path(p)
        if path.exists():
            return path

    path = Path(paths[0])
    path.mkdir(parents=True, exist_ok=True)
    return path


def patch_run_executor_source(site_packages: Path) -> bool:
    """Insert os._exit(0) after the executor writes output metadata.

    After that write, interpreter shutdown aborts in protobuf/MLMD C++
    destructors. Exiting here skips that shutdown.

    Returns True if applied or already present, False if the target or anchor is missing
    (e.g. an unexpected TFX version).
    """
    target = site_packages / _RUN_EXECUTOR_PATH
    if not target.exists():
        log.warning("run_executor not found at %s; skipping force-exit patch", target)
        return False
    source = target.read_text(encoding="utf-8")
    if _FORCE_EXIT_MARKER in source:
        log.info("force-exit source patch already present")
        return True
    if _FORCE_EXIT_ANCHOR not in source:
        log.warning(
            "force-exit anchor not found in %s; skipping (TFX version mismatch?)",
            target,
        )
        return False
    target.write_text(source.replace(_FORCE_EXIT_ANCHOR, _FORCE_EXIT_INSERT, 1), encoding="utf-8")
    log.info("Applied force-exit source patch to %s", target)
    return True


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")

    site_packages = _get_site_packages()
    target_dir = site_packages / "tfx_kfp_v2_shim"

    src_dir = Path(__file__).parent
    if target_dir.exists():
        shutil.rmtree(target_dir)
    shutil.copytree(
        src_dir,
        target_dir,
        ignore=shutil.ignore_patterns(
            "__pycache__",
            "*.pyc",
            "tests",
            "install_shim.py",
        ),
    )
    log.info("Installed shim package to %s", target_dir)

    pth_file = site_packages / "tfx_kfp_v2_shim.pth"
    pth_file.write_text(PTH_CONTENT, encoding="utf-8")
    log.info("Installed .pth file: %s", pth_file)
    log.info("Shim will activate on every Python invocation.")

    # The runtime hook cannot patch the `python -m` executor entry module, so this
    # source edit is the real fix for the shutdown crash. A failed patch must fail
    # the build loudly, otherwise the image ships without the fix and the executor
    # still aborts on shutdown.
    if not patch_run_executor_source(site_packages):
        log.error(
            "force-exit source patch was not applied to %s; refusing to install "
            "a shim that leaves the executor shutdown crash unfixed",
            site_packages / _RUN_EXECUTOR_PATH,
        )
        sys.exit(1)


if __name__ == "__main__":
    main()
