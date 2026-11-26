"""Installer tests for the shim package."""

import pytest

from tfx_kfp_v2_shim import install_shim


def _executor_source():
    return "import os\n" + install_shim._FORCE_EXIT_ANCHOR + "\n"


def test_patch_run_executor_source_applies_once_and_is_idempotent(tmp_path):
    target = tmp_path / install_shim._RUN_EXECUTOR_PATH
    target.parent.mkdir(parents=True)
    target.write_text(_executor_source())

    assert install_shim.patch_run_executor_source(tmp_path)
    patched = target.read_text()
    assert install_shim._FORCE_EXIT_MARKER in patched
    assert patched.count("os._exit(0)") == 1
    assert install_shim.patch_run_executor_source(tmp_path)
    assert target.read_text() == patched


def test_patch_run_executor_source_rejects_missing_target(tmp_path):
    assert not install_shim.patch_run_executor_source(tmp_path)


def test_patch_run_executor_source_rejects_unexpected_target(tmp_path):
    target = tmp_path / install_shim._RUN_EXECUTOR_PATH
    target.parent.mkdir(parents=True)
    target.write_text("def unrelated(): pass\n")
    assert not install_shim.patch_run_executor_source(tmp_path)


def test_get_site_packages_prefers_existing_primary_path(tmp_path, monkeypatch):
    primary = tmp_path / "primary"
    primary.mkdir()
    monkeypatch.setattr(install_shim.site, "getsitepackages", lambda: [str(primary)])
    assert install_shim._get_site_packages() == primary


def test_get_site_packages_uses_user_site_when_primary_list_empty(tmp_path, monkeypatch):
    user = tmp_path / "user"
    user.mkdir()
    monkeypatch.setattr(install_shim.site, "getsitepackages", lambda: [])
    monkeypatch.setattr(install_shim.site, "getusersitepackages", lambda: str(user))
    assert install_shim._get_site_packages() == user


def test_get_site_packages_creates_first_unavailable_path(tmp_path, monkeypatch):
    created = tmp_path / "created" / "site-packages"
    monkeypatch.setattr(install_shim.site, "getsitepackages", lambda: [str(created)])
    assert install_shim._get_site_packages() == created
    assert created.is_dir()


def test_main_copies_package_writes_pth(tmp_path, monkeypatch):
    calls = []

    def patch_source(site_packages):
        calls.append(site_packages)
        return True

    monkeypatch.setattr(install_shim, "_get_site_packages", lambda: tmp_path)
    monkeypatch.setattr(install_shim, "patch_run_executor_source", patch_source)
    install_shim.main()

    target = tmp_path / "tfx_kfp_v2_shim"
    assert (target / "__init__.py").exists()
    assert not (target / "install_shim.py").exists()
    assert (tmp_path / "tfx_kfp_v2_shim.pth").read_text() == install_shim.PTH_CONTENT
    assert calls == [tmp_path]


def test_main_fails_install_when_force_exit_patch_cannot_apply(tmp_path, monkeypatch):
    monkeypatch.setattr(install_shim, "_get_site_packages", lambda: tmp_path)
    monkeypatch.setattr(install_shim, "patch_run_executor_source", lambda path: False)
    with pytest.raises(SystemExit) as exc:
        install_shim.main()
    assert exc.value.code == 1
    assert (tmp_path / "tfx_kfp_v2_shim").is_dir()
