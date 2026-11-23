"""Import-hook tests that do not require TFX."""

import contextlib
import importlib.machinery
import sys
import types

import pytest

from tfx_kfp_v2_shim import _hook


def test_finder_finds_registered_modules_and_bypasses_recursion():
    finder = _hook.TfxShimFinder({"fake.target": lambda module: None})

    spec = finder.find_spec("fake.target")
    assert isinstance(spec, importlib.machinery.ModuleSpec)
    assert spec.origin == "tfx-kfp-v2-shim"
    assert finder.find_spec("not.registered") is None

    with finder._bypass("fake.target"):
        assert finder.find_spec("fake.target") is None
        assert "fake.target" in finder._in_progress
    assert "fake.target" not in finder._in_progress

    with pytest.raises(RuntimeError, match="exercise cleanup"):
        with finder._bypass("fake.target"):
            raise RuntimeError("exercise cleanup")
    assert not finder._in_progress


def test_register_is_idempotent_and_unregister_is_safe():
    finder = _hook.TfxShimFinder({})
    original = list(sys.meta_path)
    try:
        finder.register()
        finder.register()
        assert sys.meta_path.count(finder) == 1
    finally:
        finder.unregister()
        finder.unregister()
    assert sys.meta_path == original


def test_loader_patches_and_copies_module(monkeypatch):
    fullname = "fake.loader_target"
    missing = object()
    previous_module = sys.modules.get(fullname, missing)
    real_module = types.ModuleType(fullname)
    real_module.value = "real"
    passed_module = types.ModuleType(fullname)
    patched = []
    bypassed = []

    monkeypatch.setattr(_hook.importlib, "import_module", lambda name: real_module)

    def apply_patch(module):
        patched.append(module)
        module.value = "patched"

    def bypass(_fullname):
        bypassed.append(_fullname)
        return contextlib.nullcontext()

    loader = _hook._ShimLoader(
        fullname,
        apply_patch,
        bypass,
    )
    assert loader.create_module(None) is None
    try:
        loader.exec_module(passed_module)
        assert patched == [real_module]
        assert bypassed == [fullname]
        assert passed_module.value == "patched"
        assert sys.modules[fullname] is passed_module
    finally:
        if previous_module is missing:
            sys.modules.pop(fullname, None)
        else:
            sys.modules[fullname] = previous_module


def test_loader_get_code_delegates_to_real_loader(monkeypatch):
    expected = object()

    class RealLoader:
        def get_code(self, fullname):
            assert fullname == "fake.code_target"
            return expected

    real_spec = types.SimpleNamespace(loader=RealLoader())
    monkeypatch.setattr(_hook.importlib.util, "find_spec", lambda name: real_spec)
    loader = _hook._ShimLoader(
        "fake.code_target",
        lambda module: None,
        lambda name: contextlib.nullcontext(),
    )
    assert loader.get_code("fake.code_target") is expected
