include common.mk
include version.mk
include newline.mk
include minikube.mk
include help.mk


# Image URL to use all building/pushing image targets
IMG ?= kfp-operator-controller

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd rbac:roleName=manager-role webhook paths="./apis/..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./apis/..."

generate-grpc: ## Generate grpc services from proto files
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=.  --go-grpc_opt=paths=source_relative \
	triggers/run-completion-event-trigger/proto/run_completion_event_trigger.proto

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

git-status-check: ## Check if there are uncommitted or untracked files
	@if [ -n "$$(git status -s)" ]; then echo "Uncommitted or untracked files: "; git status -s ; exit 1; fi

decoupled-test: manifests generate ## Run decoupled acceptance tests
	go test ./controllers/pipelines/... -tags=decoupled -coverprofile cover.out
	go test ./controllers/webhook/... -tags=decoupled
	$(MAKE) -C provider-service decoupled-test

ARGO_VERSION=$(shell sed -n 's/[^ tab]*github.com\/argoproj\/argo-workflows\/v3 \(v[.0-9]*\)[^.0-9]*/\1/p' <go.mod)
integration-test-up: ## Spin up a minikube cluster for integration tests
	minikube start -p kfp-operator-tests --registry-mirror="https://mirror.gcr.io"
	# Install Argo
	kubectl create namespace argo --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/quick-start-postgres.yaml
	kubectl wait -n argo deployment/workflow-controller --for condition=available --timeout=5m
	# Proxy K8s API
	kubectl proxy --port=8080 & echo $$! > config/testing/pids

integration-test: manifests generate helm-cmd yq ## Run integration tests
	eval $$(minikube -p kfp-operator-tests docker-env) && \
	$(MAKE) -C compilers/stub docker-build && \
	$(MAKE) -C provider-service/stub docker-build && \
	kubectl apply -n argo -f config/testing/provider-deployment.yaml
	kubectl wait -n argo deployment/provider-test --for condition=available --timeout=5m
	kubectl apply -n argo -f config/testing/provider-service.yaml
	$(HELM) template helm/kfp-operator --values config/testing/integration-test-values.yaml | \
 		$(YQ) e 'select(.kind == "*WorkflowTemplate")' - | \
 		kubectl apply -f -
	go test ./controllers/pipelines/internal/workflowfactory/... -tags=integration --timeout 20m

integration-test-down: ## Tear down the minikube cluster
	(cat config/testing/pids | xargs kill) || true
	minikube delete -p kfp-operator-tests

unit-test: manifests generate ## Run unit tests
	go test ./... -tags=unit

functional-test: ## Run functional tests
	$(MAKE) -C triggers/run-completion-event-trigger functional-test

test: fmt vet unit-test decoupled-test functional-test ## Run all tests
	# TODO: after integration tests can run on CI, run provider-service as part
	# of integration-test
	$(MAKE) -C provider-service integration-test

test-compilers: ## Run all tests for compilers
	$(MAKE) -C compilers test-all

test-all: test helm-test-operator test-compilers ## Run all tests

integration-test-all: integration-test ## Run all integration tests
	$(MAKE) -C compilers integration-test-all

##@ Build

build: generate fmt vet ## Build manager binary.
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go --zap-devel --config config/manager/controller_manager_config.yaml

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

##@ Tools

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

DYFF = $(PROJECT_DIR)/bin/dyff
dyff: ## Download dyff locally if necessary.
	$(call go-install,$(DYFF),github.com/homeport/dyff/cmd/dyff@v1.4.5)

YQ = $(PROJECT_DIR)/bin/yq
yq: ## Download yq locally if necessary.
	$(call go-install,$(YQ),github.com/mikefarah/yq/v4@v4.45.1)

HELM := $(PROJECT_DIR)/bin/helm
# Can't be named helm because it's already a directory
helm-cmd: ## Download helm locally if necessary.
	$(call go-install,$(HELM),helm.sh/helm/v3/cmd/helm@v3.15.4)

CONTROLLER_GEN = $(PROJECT_DIR)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0)

KUSTOMIZE = $(PROJECT_DIR)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)

##@ Package
helm-package-operator: helm-cmd helm-test-operator ## Package and test operator helm-chart
	$(HELM) package helm/kfp-operator --version $(VERSION) --app-version $(VERSION) -d dist

helm-package: helm-package-operator ## Package operator helm-chart

helm-install-operator: helm-package-operator values.yaml ## Install operator
	$(HELM) install -f values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz

helm-uninstall-operator: ## Uninstall operator
	$(HELM) uninstall kfp-operator

helm-upgrade-operator: helm-package-operator values.yaml ## Upgrade operator with helm chart
	$(HELM) upgrade -f values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz

ifeq ($(HELM_REPOSITORIES)$(OSS_HELM_REPOSITORIES),)
helm-publish:
	$(error OSS_HELM_REPOSITORIES or HELM_REPOSITORIES must be provided as space-separated lists of URLs)
else
ifdef NETRC_FILE
helm-publish:: $(NETRC_FILE)
endif

helm-publish:: helm-package ## Publish Helm chart to repositories
	$(foreach url,$(HELM_REPOSITORIES) $(OSS_HELM_REPOSITORIES),$(call helm-upload,$(url)))

define helm-upload
@echo "Publishing Helm chart to $(1)"
@if [[ "$(1)" == "oci://"* ]]; then \
	helm push dist/kfp-operator-$(VERSION).tgz $(1)/kfp-operator; \
else \
	curl --fail --netrc-file $(NETRC_FILE) -T dist/kfp-operator-$(VERSION).tgz $(1); \
fi
$(NEWLINE)
endef
endif

INDEXED_YAML := $(YQ) e --no-doc '{([.metadata.name, .kind] | join("-")): .}'
helm-test-operator: manifests helm-cmd kustomize yq dyff ## Test operator helm chart against kustomize
	$(eval TMP := $(shell mktemp -d))

	# Create yaml files with helm and kustomize.
	$(HELM) template helm/kfp-operator -f helm/kfp-operator/test/values.yaml > $(TMP)/helm
	$(KUSTOMIZE) build config/default > $(TMP)/kustomize
	# Because both tools create multi-document files, we have to convert them into '{name}-{kind}'-indexed objects to help the diff tools
	$(INDEXED_YAML) $(TMP)/helm > $(TMP)/helm_indexed
	$(INDEXED_YAML) $(TMP)/kustomize > $(TMP)/kustomize_indexed
	$(DYFF) between --set-exit-code $(TMP)/helm_indexed $(TMP)/kustomize_indexed
	rm -rf $(TMP)

##@ Containers

include docker-targets.mk

docker-build-compilers: ## Build all pipeline framework compiler images
	$(MAKE) -C compilers docker-build-all

docker-push-compilers: ## Publish all pipeline framework compiler images
	$(MAKE) -C compilers docker-push-all

docker-build-triggers: ## Build trigger docker images
	$(MAKE) -C triggers/run-completion-event-trigger docker-build

docker-push-triggers: ## Publish trigger docker images
	$(MAKE) -C triggers/run-completion-event-trigger docker-push

docker-build-providers: ## Build provider docker images
	$(MAKE) -C provider-service docker-build

docker-push-providers: ## Publish provider docker images
	$(MAKE) -C provider-service docker-push

##@ Docs

website: ## Build website
	$(MAKE) -C docs-gen build

docker-push-quickstart: ##  Build and push quickstart docker images
	$(MAKE) -C docs-gen docker-push-quickstart

##@ Package

package-all: docker-build docker-build-compilers docker-build-triggers docker-build-providers helm-package website ## Build all packages

publish-all: docker-push docker-push-compilers docker-push-triggers docker-push-providers helm-publish ## Publish all packages

##@ CI

prBuild: test-all package-all git-status-check ## Run all tests and build all packages

cdBuild: prBuild publish-all docker-push-quickstart ## Run all tests, build all packages and publish them
