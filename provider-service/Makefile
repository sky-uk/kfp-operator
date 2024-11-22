build:
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

test:
	$(MAKE) -C base test
	$(MAKE) -C kfp test
	$(MAKE) -C vai test

##@ Containers

docker-build:
	$(MAKE) -C kfp docker-build
	$(MAKE) -C vai docker-build

docker-push:
	$(MAKE) -C kfp docker-push
	$(MAKE) -C vai docker-push


