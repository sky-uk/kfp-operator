VERSION := $$(git describe --tags)
IMG := kfp-operator-argo-kfp-sdk:$(VERSION)

all: build

##@ Development

test:
	@echo Nothing to test.

##@ Build

build:
	@echo Nothing to build.

docker-build:
	docker build . -t $(IMG)

docker-push:
	docker push $(IMG)