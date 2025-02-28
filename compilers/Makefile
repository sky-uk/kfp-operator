include ../docker-targets.mk
include ../help.mk

all: build

##@ Development

test: ## Test
	$(MAKE) -C tfx test

integration-test-all: ## Integration test
	$(MAKE) -C compilers/tfx integration-test

##@ Build

build: ## Build
	$(MAKE) -C tfx build

##@ Containers

docker-build-all: ## Push all compiler docker images
	$(MAKE) -C tfx docker-build

docker-push-all: ## Push all compiler docker images
	$(MAKE) -C tfx docker-push
