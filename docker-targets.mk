_DOCKER_TARGETS_MK_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

include $(_DOCKER_TARGETS_MK_DIR)/version.mk
include $(_DOCKER_TARGETS_MK_DIR)/newline.mk

ifeq ($(CONTAINER_REPOSITORIES)$(OSS_CONTAINER_REGISTRY_HOSTS),)
docker-push:
	$(error CONTAINER_REPOSITORIES or OSS_CONTAINER_REGISTRY_HOSTS must be provided as a space-separated lists of hosts/registry urls)
else ifneq ($(VERSION), $(VERSION:-dirty=))
docker-push:
	$(error Refusing to push dirty image $(VERSION))
else
docker-push: docker-build ## Push container image
	$(foreach host,$(CONTAINER_REPOSITORIES) $(OSS_CONTAINER_REGISTRY_HOSTS),$(call docker-push-to-registry,$(host)))
define docker-push-to-registry
	docker tag ${IMG} $(1)/${IMG}$(NEWLINE)
	docker push $(1)/${IMG}$(NEWLINE)

	docker tag ${IMG}:${VERSION} $(1)/${IMG}:${VERSION}$(NEWLINE)
	docker push $(1)/${IMG}:${VERSION}$(NEWLINE)
endef
endif

docker-build: GOOS=linux
docker-build: GOARCH=amd64
docker-build: build ## Build container image
	docker build ${DOCKER_BUILD_EXTRA_PARAMS} --platform ${GOOS}/${GOARCH} -t ${IMG} -t ${IMG}:${VERSION} -f Dockerfile .
