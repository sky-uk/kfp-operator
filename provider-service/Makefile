build:
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

test:
#	$(MAKE) -C base test
	$(MAKE) -C kfp test
	$(MAKE) -C vai test

#functional-test: docker-build
#	docker compose up -d
#	go test ./... -tags=functional
#	docker compose down

##@ Containers

docker-build:
	$(MAKE) -C kfp docker-build
	$(MAKE) -C vai docker-build

docker-push:
	$(MAKE) -C kfp docker-push
	$(MAKE) -C vai docker-push


