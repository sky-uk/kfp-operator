"""Tests for the TFX/KFP v2 shim package initializer."""

import pytest

import tfx_kfp_v2_shim as shim


@pytest.fixture
def recorded_finder(monkeypatch: pytest.MonkeyPatch):
    created = []

    class FakeFinder:
        def __init__(self, patches):
            self.patches = patches
            created.append(self)

        def register(self):
            self.registered = True

    monkeypatch.setattr(shim, "TfxShimFinder", FakeFinder)
    monkeypatch.setattr(shim, "_installed", False)
    return created


def test_install_registers_expected_patches(recorded_finder):
    shim.install()

    expected = {
        shim._COMPILER_UTILS: shim.patch_compiler_utils,
        shim._ENTRYPOINT_UTILS: shim.patch_entrypoint_utils,
        shim._EXPERIMENTAL: shim.patch_experimental,
        shim._RUN_EXECUTOR: shim.patch_run_executor,
    }

    finder = recorded_finder[0]
    assert finder.registered
    assert finder.patches == expected


def test_install_is_idempotent(recorded_finder):
    shim.install()
    shim.install()
    assert len(recorded_finder) == 1
