"""Make TFX run correctly on Kubeflow Pipelines v2.

install() registers an import hook that patches TFX modules as they load.

Usage:
    import tfx_kfp_v2_shim
    tfx_kfp_v2_shim.install()
"""

from collections.abc import Callable
from types import ModuleType

from tfx_kfp_v2_shim._hook import TfxShimFinder
from tfx_kfp_v2_shim._patches import (
    patch_compiler_utils,
    patch_entrypoint_utils,
    patch_experimental,
    patch_run_executor,
)

__all__ = ["install"]

# Target TFX module paths
_COMPILER_UTILS = "tfx.orchestration.kubeflow.v2.compiler_utils"
_ENTRYPOINT_UTILS = "tfx.orchestration.kubeflow.v2.container.kubeflow_v2_entrypoint_utils"
_EXPERIMENTAL = "tfx.v1.orchestration.experimental"
_RUN_EXECUTOR = "tfx.orchestration.kubeflow.v2.container.kubeflow_v2_run_executor"

_installed = False


def install() -> None:
    """Register the import hook once on sys.meta_path."""
    global _installed
    if _installed:
        return

    patches: dict[str, Callable[[ModuleType], None]] = {
        _COMPILER_UTILS: patch_compiler_utils,
        _ENTRYPOINT_UTILS: patch_entrypoint_utils,
        _EXPERIMENTAL: patch_experimental,
        _RUN_EXECUTOR: patch_run_executor,
    }

    finder = TfxShimFinder(patches)
    finder.register()
    _installed = True
