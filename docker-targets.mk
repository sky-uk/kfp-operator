ifdef CONTAINER_REGISTRY_HOSTS
docker-push: docker-build $(CONTAINER_REGISTRY_HOSTS) ## Push container image
$(CONTAINER_REGISTRY_HOSTS):
	docker tag ${IMG}:${VERSION} $@/${IMG}:${VERSION}
	docker push $@/${IMG}:${VERSION}
else
docker-push:
	$(error CONTAINER_REGISTRY_HOSTS must be a space-separated list of hosts)
endif

docker-build: build ## Build container image
	docker build -t ${IMG} -t ${IMG}:${VERSION} .
