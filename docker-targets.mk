ifndef CONTAINER_REGISTRY_HOSTS
docker-push:
	$(error CONTAINER_REGISTRY_HOSTS must be a space-separated list of hosts)
else ifneq ($(VERSION), $(VERSION:-dirty=))
docker-push:
	$(error Refusing to push dirty image $(VERSION))
else
docker-push: docker-build $(CONTAINER_REGISTRY_HOSTS) ## Push container image
$(CONTAINER_REGISTRY_HOSTS):
	docker tag ${IMG}:${VERSION} $@/${IMG}:${VERSION}
	docker push $@/${IMG}:${VERSION}
endif

docker-build: build ## Build container image
	docker build ${DOCKER_BUILD_EXTRA_PARAMS} -t ${IMG} -t ${IMG}:${VERSION} .
