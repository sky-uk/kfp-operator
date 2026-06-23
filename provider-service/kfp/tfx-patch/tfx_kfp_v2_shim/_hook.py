"""PEP 451 import hook that intercepts TFX modules and patches them at load time."""

from __future__ import annotations

import importlib
import importlib.abc
import importlib.machinery
import importlib.util
import logging
import sys
from typing import Callable, Dict, Optional, Sequence

log = logging.getLogger(__name__)


class TfxShimFinder(importlib.abc.MetaPathFinder):
    """Meta-path finder: intercepts registered modules and delegates to _ShimLoader."""

    def __init__(self, patches: Dict[str, Callable]) -> None:
        self._patches = dict(patches)
        self._in_progress: set[str] = set()

    def find_spec(self, fullname: str, path: Optional[Sequence[str]] = None, target=None):
        if fullname in self._patches and fullname not in self._in_progress:
            return importlib.machinery.ModuleSpec(
                fullname, _ShimLoader(self, fullname), origin="tfx-kfp-v2-shim",
            )
        return None

    def register(self) -> None:
        if self not in sys.meta_path:
            sys.meta_path.insert(0, self)
            log.info("TfxShimFinder registered for %s", list(self._patches))

    def unregister(self) -> None:
        try:
            sys.meta_path.remove(self)
        except ValueError:
            pass


class _ShimLoader(importlib.abc.Loader):
    """Loads the real module, patches it, then installs the patched version."""

    def __init__(self, finder: TfxShimFinder, fullname: str) -> None:
        self._finder = finder
        self._fullname = fullname

    def create_module(self, spec):
        return None

    def exec_module(self, module) -> None:
        fullname = self._fullname
        sys.modules.pop(fullname, None)

        self._finder._in_progress.add(fullname)
        try:
            real_module = importlib.import_module(fullname)
        finally:
            self._finder._in_progress.discard(fullname)

        # Patch the real module first — its functions hold __globals__ refs
        # to real_module.__dict__, so patches must land there.
        self._finder._patches[fullname](real_module)
        log.info("Patched %s", fullname)

        module.__dict__.update(real_module.__dict__)
        sys.modules[fullname] = module

    def get_code(self, fullname: str):
        """Required by runpy (``python -m``)."""
        self._finder._in_progress.add(fullname)
        try:
            real_spec = importlib.util.find_spec(fullname)
        finally:
            self._finder._in_progress.discard(fullname)

        if real_spec and real_spec.loader and hasattr(real_spec.loader, "get_code"):
            return real_spec.loader.get_code(fullname)
        return None
