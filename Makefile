include common.mk
include version.mk
include newline.mk

# Image URL to use all building/pushing image targets
IMG ?= kfp-operator-controller
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:preserveUnknownFields=false"

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

git-status-check:
	@if [ -n "$$(git status -s)" ]; then echo "Uncommitted or untracked files: "; git status -s ; exit 1; fi

decoupled-test: manifests generate ## Run decoupled acceptance tests
	$(call envtest-run,go test ./... -tags=decoupled -coverprofile cover.out)

ARGO_VERSION=$(shell sed -n 's/[^ tab]*github.com\/argoproj\/argo-workflows\/v3 \(v[.0-9]*\)[^.0-9]*/\1/p' <go.mod)
integration-test-up:
	minikube start -p kfp-operator-tests
	# Install Argo
	kubectl create namespace argo --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/quick-start-postgres.yaml
	kubectl wait -n argo deployment/workflow-controller --for condition=available --timeout=5m
	# Proxy K8s API
	kubectl proxy --port=8080 & echo $$! > config/testing/pids

integration-test: manifests generate helm-cmd yq ## Run integration tests
	eval $$(minikube -p kfp-operator-tests docker-env) && \
	$(MAKE) -C argo/providers/stub docker-build && \
	docker build docs-gen/includes/quickstart -t kfp-quickstart
	$(HELM) template helm/kfp-operator --values config/testing/integration-test-values.yaml | \
 		$(YQ) e 'select(.kind == "*WorkflowTemplate")' - | \
 		kubectl apply -f -
	go test ./... -tags=integration --timeout 20m

integration-test-down:
	(cat config/testing/pids | xargs kill) || true
	minikube stop -p kfp-operator-tests

unit-test: manifests generate ## Run unit tests
	go test ./... -tags=unit

test: fmt vet unit-test decoupled-test

test-argo:
	$(MAKE) -C argo/common test
	$(MAKE) -C argo/status-updater test
	#$(MAKE) -C argo/kfp-compiler test
	$(MAKE) -C argo/providers test

test-all: test helm-test test-argo

integration-test-all: integration-test
	$(MAKE) -C argo/kfp-compiler integration-test

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
	$(call go-install,$(YQ),github.com/mikefarah/yq/v4@v4.13.2)

HELM := $(PROJECT_DIR)/bin/helm
# Can't be named helm because it's already a directory
helm-cmd: ## Download helm locally if necessary.
	$(call go-install,$(HELM),helm.sh/helm/v3/cmd/helm@v3.7.0)

CONTROLLER_GEN = $(PROJECT_DIR)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(PROJECT_DIR)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)

##@ Package

helm-package: helm-cmd helm-test
	$(HELM) package helm/kfp-operator --version $(VERSION) --app-version $(VERSION) -d dist

helm-install: helm-package values.yaml
	$(HELM) install -f values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz

helm-uninstall:
	$(HELM) uninstall kfp-operator

helm-upgrade: helm-package values.yaml
	$(HELM) upgrade -f values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz

ifeq ($(HELM_REPOSITORIES)$(OSS_HELM_REPOSITORIES),)
helm-publish:
	$(error OSS_HELM_REPOSITORIES or HELM_REPOSITORIES must be provided as space-separated lists of URLs)
else
ifdef NETRC_FILE
helm-publish:: $(NETRC_FILE)
endif

helm-publish:: helm-package
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

INDEXED_YAML := $(YQ) e '{([.metadata.name, .kind] | join("-")): .}'
helm-test: manifests helm-cmd kustomize yq dyff
	$(eval TMP := $(shell mktemp -d))

	# Create yaml files with helm and kustomize.
	$(HELM) template helm/kfp-operator -f helm/kfp-operator/test/values.yaml > $(TMP)/helm
	$(KUSTOMIZE) build config/default > $(TMP)/kustomize
	# Because both tools create multi-document files, we have to convert them into '{kind}-{name}'-indexed objects to help the diff tools
	$(INDEXED_YAML) $(TMP)/helm > $(TMP)/helm_indexed
	$(INDEXED_YAML) $(TMP)/kustomize > $(TMP)/kustomize_indexed
	$(DYFF) between --set-exit-code $(TMP)/helm_indexed $(TMP)/kustomize_indexed
	rm -rf $(TMP)

##@ Containers

include docker-targets.mk

docker-build-argo:
	$(MAKE) -C argo/status-updater docker-build
	#$(MAKE) -C argo/kfp-compiler docker-build
	$(MAKE) -C argo/providers docker-build

docker-push-argo:
	$(MAKE) -C argo/status-updater docker-push
	#$(MAKE) -C argo/kfp-compiler docker-push
	$(MAKE) -C argo/providers docker-push

##@ Docs
website:
	$(MAKE) -C docs-gen

docker-push-quickstart:
	$(MAKE) -C docs-gen/includes/quickstart docker-push

##@ Package

package-all: docker-build docker-build-argo helm-package website

publish-all: docker-push docker-push-argo helm-publish

##@ CI

prBuild: test-all package-all git-status-check

cdBuild: prBuild publish-all docker-push-quickstart
