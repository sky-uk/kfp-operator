ifndef go-install

# go-install will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
export PATH := $(PATH):$(PROJECT_DIR)/bin

define go-install
GOBIN=$(PROJECT_DIR)/bin go install $(2)
endef

endif # go-install
