ifndef commons-include
define commons-include
endef

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
export PATH := $(PATH):$(PROJECT_DIR)/bin

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

ENVTEST_ASSETS_DIR=$(PROJECT_DIR)/testbin
CONTROLLER_RUNTIME_VERSION := $(shell go list -m -f '{{.Version}}' sigs.k8s.io/controller-runtime)

define envtest-run
	mkdir -p ${ENVTEST_ASSETS_DIR}; \
    test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || \
	curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/$(CONTROLLER_RUNTIME_VERSION)/hack/setup-envtest.sh; \
    source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); $(1)
endef

# go-install will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
export PATH := $(PATH):$(PROJECT_DIR)/bin

define go-install
GOBIN=$(PROJECT_DIR)/bin go install $(2)
endef

endif # commons-include
