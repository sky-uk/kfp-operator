include ../../../common.mk
include ../../../docker-targets.mk
include ../../../help.mk

IMG := kfp-operator-tfx-patch

##@ Development

test: ## Run unit tests for the shim package
	python -m pytest tests/ -v

##@ Build

build: ## No-op — patch image has no build step
