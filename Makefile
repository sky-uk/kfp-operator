include docker.mk

# Image URL to use all building/pushing image targets
IMG ?= kfp-operator-controller
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
decoupled-test: ## Run decoupled acceptance tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -tags=decoupled -coverprofile cover.out

integration-test-up:
	minikube start -p argo-integration-tests
	kubectl create namespace argo --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo-workflows/master/manifests/quick-start-postgres.yaml
	kubectl apply -n argo -f config/testing/wiremock.yaml
	rm -f config/testing/pids
	kubectl wait -n argo deployment/wiremock --for condition=available --timeout=5m
	kubectl port-forward -n argo service/wiremock 8081:80 & echo $$! >> config/testing/pids
	kubectl wait -n argo deployment/workflow-controller --for condition=available --timeout=5m
	kubectl proxy --port=8080 & echo $$! >> config/testing/pids

integration-test: ## Run integration tests
	eval $$(minikube -p argo-integration-tests docker-env) && \
	$(MAKE) -C argo/compiler docker-build && \
	$(MAKE) -C argo/kfp-sdk docker-build && \
	docker build docs/quickstart -t kfp-quickstart
	go test ./... -tags=integration

integration-test-down:
	(cat config/testing/pids | xargs kill) || true
	minikube stop -p argo-integration-tests

unit-test: ## Run unit tests
	go test ./... -tags=unit

test: manifests generate fmt vet unit-test # decoupled-test

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go --config config/manager/controller_manager_config.yaml

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

##@ Helm

helm-package:
	$(HELM) package config/helm/kfp-operator --version $(VERSION) --app-version $(VERSION) -d dist

helm-install: helm-package values.yaml
	$(HELM) install -f values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz

helm-uninstall:
	$(HELM) uninstall kfp-operator

helm-test: manifests helm kustomize yq dyff
	$(eval TMP := $(shell mktemp -d))

	# Create yaml files with helm and kustomize.
	$(HELM) template config/helm/kfp-operator -f config/helm/kfp-operator/test/values.yaml > $(TMP)/helm
	$(KUSTOMIZE) build config/default > $(TMP)/kustomize
	# Because both tools create multi-document files, we have to convert them into '{kind}-{name}'-indexed objects to help the diff tools
	$(INDEXED_YAML) $(TMP)/helm > $(TMP)/helm_indexed
	$(INDEXED_YAML) $(TMP)/kustomize > $(TMP)/kustomize_indexed
	$(DYFF) between --set-exit-code $(TMP)/helm_indexed $(TMP)/kustomize_indexed
	rm -rf $(TMP)

##@ Tools
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

DYFF = $(PROJECT_DIR)/bin/dyff
dyff: ## Download yaml-diff locally if necessary.
	$(call go-get-tool,$(DYFF),github.com/homeport/dyff/cmd/dyff@v1.4.5)

YQ = $(PROJECT_DIR)/bin/yq
yq: ## Download yaml-diff locally if necessary.
	$(call go-get-tool,$(YQ),github.com/mikefarah/yq/v4@v4.13.2)
INDEXED_YAML := $(YQ) e '{([.metadata.name, .kind] | join("-")): .}'

HELM := $(PROJECT_DIR)/bin/helm
helm: ## Download helm locally if necessary.
	$(call go-get-tool,$(HELM),helm.sh/helm/v3/cmd/helm@v3.7.0)

CONTROLLER_GEN = $(PROJECT_DIR)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(PROJECT_DIR)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

##@ Docker

docker-build-argo:
	$(MAKE) -C argo/compiler docker-build # test
	$(MAKE) -C argo/kfp-sdk test docker-build

docker-push-argo:
	$(MAKE) -C argo/compiler docker-push
	$(MAKE) -C argo/kfp-sdk docker-push

##@ CI

prBuild: test helm-test docker-build docker-build-argo # decoupled-test


cdBuild: prBuild docker-push docker-push-argo
