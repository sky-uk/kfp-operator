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
	DOCKER_BUILDKIT=1 docker build ${DOCKER_BUILD_EXTRA_PARAMS} \
		--platform ${GOOS}/${GOARCH} \
		--cache-from type=registry,ref=$(firstword $(CONTAINER_REPOSITORIES) $(OSS_CONTAINER_REGISTRY_HOSTS))/${IMG}:cache \
		--cache-to   type=registry,ref=$(firstword $(CONTAINER_REPOSITORIES) $(OSS_CONTAINER_REGISTRY_HOSTS))/${IMG}:cache,mode=max \
		-t ${IMG} -t ${IMG}:${VERSION} -f Dockerfile .
