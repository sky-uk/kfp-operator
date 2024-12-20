build:
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

unit-test:
	$(MAKE) -C base unit-test
	$(MAKE) -C kfp unit-test
	$(MAKE) -C vai unit-test

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


