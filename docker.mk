VERSION := $(shell git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*')

define docker_push
	for HOST in $(CONTAINER_REGISTRY_HOSTS); \
    do \
    	TAG="$$HOST/$(1):${VERSION}"; \
    	docker tag $(1) $$TAG; \
    	docker push $$TAG; \
    done
endef
