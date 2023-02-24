_DOCKER_TARGETS_MK_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include $(_DOCKER_TARGETS_MK_DIR)/version.mk
include $(_DOCKER_TARGETS_MK_DIR)/newline.mk

ifeq ($(CONTAINER_REGISTRY_HOSTS)$(OSS_CONTAINER_REGISTRY_HOSTS),)
docker-push:
	$(error CONTAINER_REGISTRY_HOSTS or OSS_CONTAINER_REGISTRY_HOSTS must be provided as a space-separated lists of hosts)
else ifneq ($(VERSION), $(VERSION:-dirty=))
docker-push:
	$(error Refusing to push dirty image $(VERSION))
else
docker-push: docker-build ## Push container image
	$(foreach host,$(CONTAINER_REGISTRY_HOSTS) $(OSS_CONTAINER_REGISTRY_HOSTS),$(call docker-push-to-registry,$(host)))
define docker-push-to-registry
	docker tag ${IMG}:${VERSION} $(1)/${IMG}:${VERSION}$(NEWLINE)
	docker push $(1)/${IMG}:${VERSION}$(NEWLINE)
endef
endif

docker-build: build ## Build container image
	docker build ${DOCKER_BUILD_EXTRA_PARAMS} -t ${IMG} -t ${IMG}:${VERSION} -f Dockerfile $(_DOCKER_TARGETS_MK_DIR)
