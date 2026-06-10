"""Import hook (PEP 451) that intercepts TFX module imports and applies patches.

The ``TfxShimFinder`` sits on ``sys.meta_path``.  When one of the registered
target modules is imported it:

  1. Temporarily disables itself (re-entrancy guard) so the *real* finder can
     locate the module.
  2. Lets the real loader populate the module namespace.
  3. Calls the registered patch function to monkey-patch the module in-place.
  4. Puts the patched module back into ``sys.modules``.
"""

from __future__ import annotations

import importlib
import importlib.abc
import importlib.machinery
import logging
import sys
from typing import Callable, Dict, Optional, Sequence

log = logging.getLogger(__name__)


class TfxShimFinder(importlib.abc.MetaPathFinder):
    """Meta-path finder that intercepts target TFX modules and patches them."""

    def __init__(self, patches: Dict[str, Callable]) -> None:
        self._patches = dict(patches)
        self._in_progress: set[str] = set()

    # -- MetaPathFinder interface ------------------------------------------

    def find_spec(
        self,
        fullname: str,
        path: Optional[Sequence[str]] = None,
        target=None,
    ):
        if fullname in self._patches and fullname not in self._in_progress:
            return importlib.machinery.ModuleSpec(
                fullname,
                _ShimLoader(self, fullname),
                origin="tfx-kfp-v2-shim",
            )
        return None

    # -- public API --------------------------------------------------------

    def register(self) -> None:
        """Insert this finder at the front of ``sys.meta_path``."""
        if self not in sys.meta_path:
            sys.meta_path.insert(0, self)
            log.info("TfxShimFinder registered for %s", list(self._patches))

    def unregister(self) -> None:
        """Remove this finder from ``sys.meta_path``."""
        try:
            sys.meta_path.remove(self)
        except ValueError:
            pass


class _ShimLoader(importlib.abc.Loader):
    """Loader that delegates to the real loader then applies a patch."""

    def __init__(self, finder: TfxShimFinder, fullname: str) -> None:
        self._finder = finder
        self._fullname = fullname

    def create_module(self, spec):
        return None  # use default module creation

    def exec_module(self, module) -> None:
        fullname = self._fullname

        # Remove our placeholder from sys.modules so the real import can run.
        sys.modules.pop(fullname, None)

        self._finder._in_progress.add(fullname)
        try:
            real_module = importlib.import_module(fullname)
        finally:
            self._finder._in_progress.discard(fullname)

        # Copy the real module's namespace into the module object that the
        # import system gave us, so callers get a single consistent object.
        module.__dict__.update(real_module.__dict__)

        # Put *our* module back (the import system expects it there).
        sys.modules[fullname] = module

        # Apply the patch.
        patch_fn = self._finder._patches[fullname]
        patch_fn(module)
        log.info("Patched %s via import hook", fullname)
