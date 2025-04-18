minikube-install-dependencies:
	$(HELM) repo add argo https://argoproj.github.io/argo-helm
	$(HELM) install argo-workflows argo/argo-workflows -n argo --create-namespace
	$(HELM) install argo-events argo/argo-events -n argo-events --create-namespace
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.crds.yaml
	openssl req -new -newkey rsa:2048 -days 365 -keyout ./local/kfp-operator-webhook.key -out ./local/kfp-operator-webhook.csr \
	  -subj "/C=US/ST=California/L=San Francisco/O=My Organization/OU=My Unit/CN=kfp-operator-webhook-service.kfp-operator-system.svc" \
	  -extensions san -config <(echo "[req]"; echo "distinguished_name=req_distinguished_name"; echo "x509_extensions = san"; \
	  echo "[req_distinguished_name]"; echo "C=US"; echo "ST=California"; echo "L=San Francisco"; echo "O=My Organization"; \
	  echo "OU=My Unit"; echo "CN=kfp-operator-webhook-service.kfp-operator-system.svc"; \
	  echo "[san]"; echo "subjectAltName=DNS:kfp-operator-webhook-service.kfp-operator-system.svc,DNS:kfp-operator-webhook-service") -nodes
	openssl x509 -req -in ./local/kfp-operator-webhook.csr -signkey ./local/kfp-operator-webhook.key -out ./local/kfp-operator-webhook.crt \
	  -extensions v3_req -extfile <(echo "[v3_req]"; echo "subjectAltName=DNS:kfp-operator-webhook-service.kfp-operator-system.svc,DNS:kfp-operator-webhook-service")
	kubectl create namespace kfp-operator-system
	kubectl create namespace vai
	kubectl create namespace kfp
	kubectl create secret tls webhook-server-cert --cert=./local/kfp-operator-webhook.crt --key=./local/kfp-operator-webhook.key --namespace kfp-operator-system
	sleep 20

minikube-helm-install-operator: helm-package-operator ./local/values.yaml

	$(HELM) install -f ./local/values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz --set containerRegistry=localhost:5000/kfp-operator

minikube-helm-upgrade-operator: helm-package-operator ./local/values.yaml
	$(HELM) upgrade -f ./local/values.yaml kfp-operator dist/kfp-operator-$(VERSION).tgz --set containerRegistry=localhost:5000/kfp-operator

minikube-install-operator: export VERSION=$(shell (git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*') | sed 's/^v//')
minikube-install-operator: export REGISTRY_PORT=$(shell docker inspect local-kfp-operator --format '{{ (index .NetworkSettings.Ports "5000/tcp" 0).HostPort }}')
minikube-install-operator: export CONTAINER_REPOSITORIES=localhost:${REGISTRY_PORT}/kfp-operator
minikube-install-operator:
	$(MAKE) docker-push docker-push-triggers docker-push-providers
	$(MAKE) minikube-helm-install-operator VERSION=${VERSION} CONTAINER_REPOSITORIES=${CONTAINER_REPOSITORIES}

minikube-install-provider: export VERSION=$(shell (git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*') | sed 's/^v//')
minikube-install-provider: export REGISTRY_PORT=$(shell docker inspect local-kfp-operator --format '{{ (index .NetworkSettings.Ports "5000/tcp" 0).HostPort }}')
minikube-install-provider: export CONTAINER_REPOSITORIES=localhost:${REGISTRY_PORT}/kfp-operator
minikube-install-provider:
	$(MAKE) -C provider-service docker-push
	$(MAKE) minikube-provider-setup

minikube-provider-setup:
	@if [ -f ./provider-setup.sh ]; then \
		echo "Running provider setup script"; \
		bash ./provider-setup.sh; \
	else \
		echo "Provider setup script not found"; \
	fi

minikube-provider-teardown:
	@if [ -f ./provider-teardown.sh ]; then \
		echo "Running provider teardown script"; \
		bash ./provider-teardown.sh; \
	else \
		echo "Provider teardown script not found"; \
	fi

minikube-upgrade-operator: export VERSION=$(shell (git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*') | sed 's/^v//')
minikube-upgrade-operator: export REGISTRY_PORT=$(shell docker inspect local-kfp-operator --format '{{ (index .NetworkSettings.Ports "5000/tcp" 0).HostPort }}')
minikube-upgrade-operator: export CONTAINER_REPOSITORIES=localhost:${REGISTRY_PORT}/kfp-operator
minikube-upgrade-operator:
	$(MAKE) docker-push docker-push-triggers
	$(MAKE) minikube-helm-upgrade-operator VERSION=${VERSION} CONTAINER_REPOSITORIES=${CONTAINER_REPOSITORIES}

minikube-upgrade-provider: minikube-provider-teardown minikube-install-provider

minikube-start:
	minikube start -p local-kfp-operator --driver=docker --registry-mirror="https://mirror.gcr.io"
	minikube addons enable registry -p local-kfp-operator

minikube-up:
	$(MAKE) minikube-start
	$(MAKE) minikube-install-dependencies
	$(MAKE) minikube-install-operator
	$(MAKE) minikube-install-provider

minikube-down:
	$(MAKE) minikube-provider-teardown
	minikube delete -p local-kfp-operator
