MINIKUBE_PROFILE := local-kfp-operator
MINIKUBE_REGISTRY := localhost:5000/kfp-operator
MINIKUBE_VERSION := $(shell git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*' | sed 's/^v//')
MINIKUBE_REGISTRY_PORT = $(shell docker inspect $(MINIKUBE_PROFILE) --format '{{ (index .NetworkSettings.Ports "5000/tcp" 0).HostPort }}')
MINIKUBE_GOARCH := $(shell go env GOARCH)

##@ Local development with stub provider (no real training infrastructure needed)

minikube-start: ## Start minikube cluster with registry
	minikube start -p $(MINIKUBE_PROFILE) --driver=docker --registry-mirror="https://mirror.gcr.io"
	minikube addons enable registry -p $(MINIKUBE_PROFILE) --images="KubeRegistryProxy=gcr.io/google_containers/kube-registry-proxy:0.4" ## The default gcr.io/k8s-minikube/kube-registry-proxy:0.0.5 isn't available anymore, the real fix is to update to the latest version of minikube
	minikube ssh -p $(MINIKUBE_PROFILE) "sudo sysctl fs.inotify.max_user_watches=524288 && sudo sysctl fs.inotify.max_user_instances=512"

minikube-install-dependencies: helm-cmd ## Install Argo Workflows, Argo Events, and cert-manager
	$(HELM) repo add argo https://argoproj.github.io/argo-helm
	$(HELM) repo add jetstack https://charts.jetstack.io
	$(HELM) repo update
	$(HELM) upgrade --install argo-workflows argo/argo-workflows -n argo --create-namespace
	$(HELM) upgrade --install argo-events argo/argo-events -n argo-events --create-namespace
	$(HELM) upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set crds.enabled=true
	kubectl create namespace kfp-operator-system --dry-run=client -o yaml | kubectl apply -f -
	@echo "Waiting for cert-manager webhook to be ready..."
	kubectl rollout status deployment/cert-manager-webhook -n cert-manager

minikube-install-minio: ## Deploy MinIO for Argo artifact storage
	@echo "Deploying MinIO for Argo artifact storage..."
	kubectl apply -f ./local/minio.yaml
	kubectl rollout status deployment/minio -n argo
	@echo "Creating argo-artifacts bucket..."
	kubectl delete pod minio-setup -n argo --ignore-not-found
	kubectl run minio-setup --image=minio/mc:latest --restart=Never -n argo --command -- \
		sh -c 'mc alias set local http://minio:9000 minioadmin minioadmin && mc mb local/argo-artifacts --ignore-existing'
	kubectl wait --for=jsonpath='{.status.phase}'=Succeeded pod/minio-setup -n argo
	kubectl delete pod minio-setup -n argo --ignore-not-found
	@echo "Configuring Argo workflow controller to use MinIO..."
	kubectl create configmap argo-workflows-workflow-controller-configmap -n argo \
		--from-file=config=./local/argo-artifact-config.yaml --dry-run=client -o yaml | kubectl apply -f -
	kubectl rollout restart deployment/argo-workflows-workflow-controller -n argo
	kubectl rollout status deployment/argo-workflows-workflow-controller -n argo
	@echo "MinIO artifact storage configured successfully."

minikube-helm-install: helm-package-operator ./local/values.yaml ## Helm install the operator
	$(HELM) upgrade --install -f ./local/values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz --set containerRegistry=$(MINIKUBE_REGISTRY)

minikube-helm-upgrade: helm-package-operator ./local/values.yaml ## Helm upgrade the operator
	$(HELM) upgrade -f ./local/values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz --set containerRegistry=$(MINIKUBE_REGISTRY)

minikube-install-operator: export CONTAINER_REPOSITORIES=localhost:$(MINIKUBE_REGISTRY_PORT)/kfp-operator
minikube-install-operator: ## Build and push operator + stub images, then helm install
	$(MAKE) docker-push docker-push-triggers VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) -C provider-service/stub docker-push VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) -C compilers/stub docker-push VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) minikube-helm-install VERSION=$(MINIKUBE_VERSION) CONTAINER_REPOSITORIES=${CONTAINER_REPOSITORIES}

minikube-apply-provider: ## Apply the stub Provider CR into the cluster
	@sed "s|:VERSION|:$(MINIKUBE_VERSION)|g" ./local/provider.yaml | kubectl apply -f -

minikube-install-eventing: ## Deploy EventSource and Sensor for run completion events
	@echo "Deploying Argo Events EventSource and Sensor..."
	kubectl apply -f ./local/eventsource.yaml
	kubectl apply -f ./local/sensor.yaml
	@echo "Waiting for EventBus to be ready..."
	kubectl wait --for=condition=Deployed eventbus/default -n kfp-operator-system
	@echo "Eventing resources deployed."

minikube-up: ## Spin up the full local stack from scratch
	$(MAKE) minikube-start
	$(MAKE) minikube-install-dependencies
	$(MAKE) minikube-install-operator
	$(MAKE) minikube-install-minio
	$(MAKE) minikube-install-eventing
	$(MAKE) minikube-apply-provider

minikube-upgrade: export CONTAINER_REPOSITORIES=localhost:$(MINIKUBE_REGISTRY_PORT)/kfp-operator
minikube-upgrade: ## Rebuild and upgrade operator + stub provider in-place
	$(MAKE) docker-push docker-push-triggers VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) -C provider-service/stub docker-push VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) -C compilers/stub docker-push VERSION=$(MINIKUBE_VERSION) GOARCH=$(MINIKUBE_GOARCH)
	$(MAKE) minikube-helm-upgrade VERSION=$(MINIKUBE_VERSION) CONTAINER_REPOSITORIES=$(CONTAINER_REPOSITORIES)
	$(MAKE) minikube-apply-provider

minikube-down: ## Tear down the minikube cluster
	minikube delete -p $(MINIKUBE_PROFILE)
