"""TFX ↔ KFP v2 compatibility shim.

Installs an import hook (PEP 302/451) that intercepts TFX module imports
and applies targeted monkey-patches for OSS Kubeflow Pipelines v2 compat.

Usage:
    import tfx_kfp_v2_shim
    tfx_kfp_v2_shim.install()              # all patches
    tfx_kfp_v2_shim.install(vertex_compatible=True)  # skip patch 5
"""

from tfx_kfp_v2_shim._hook import TfxShimFinder
from tfx_kfp_v2_shim._patches import (
    patch_compiler_utils,
    patch_entrypoint_utils,
    patch_experimental,
    patch_path_utils,
)

# Target TFX module paths
_COMPILER_UTILS = "tfx.orchestration.kubeflow.v2.compiler_utils"
_ENTRYPOINT_UTILS = (
    "tfx.orchestration.kubeflow.v2.container.kubeflow_v2_entrypoint_utils"
)
_PATH_UTILS = "tfx.utils.path_utils"
_EXPERIMENTAL = "tfx.v1.orchestration.experimental"

_installed = False


def install(*, vertex_compatible: bool = False) -> None:
    """Register the import hook on sys.meta_path.

    Args:
        vertex_compatible: If True, skip patch 5 (flatten model directory)
            which breaks Vertex AI compatibility.
    """
    global _installed
    if _installed:
        return

    patches = {
        _COMPILER_UTILS: patch_compiler_utils,
        _ENTRYPOINT_UTILS: patch_entrypoint_utils,
        _EXPERIMENTAL: patch_experimental,
    }
    if not vertex_compatible:
        patches[_PATH_UTILS] = patch_path_utils

    finder = TfxShimFinder(patches)
    finder.register()
    _installed = True
