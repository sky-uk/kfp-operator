VERSION := $(shell git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*' || echo 0.0.0)

##@ Container

ifdef CONTAINER_REGISTRY_HOSTS
PUSH_TARGETS := $(foreach h,${CONTAINER_REGISTRY_HOSTS},push_to_${h})
docker-push: docker-build ${PUSH_TARGETS} ## Push container image
$(PUSH_TARGETS):
	docker tag ${IMG} $@$/${IMG}:${VERSION}
	docker push $@/${IMG}:${VERSION}
else
docker-push:
	$(error CONTAINER_REGISTRY_HOSTS must be a space-separated list of hosts)
endif

docker-build: build ## Build container image
	docker build -t ${IMG} .

