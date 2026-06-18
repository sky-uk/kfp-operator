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
