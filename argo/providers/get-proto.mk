PROTOPATH := $(HOME)/.proto

define get-proto
	mkdir -p $(PROTOPATH) && \
	cd $(PROTOPATH); \
	[ -d "$(1)@$(2)" ] || { \
		git clone --no-checkout --branch $(2) --single-branch https://$(1) $(1)@$(2) && \
		cd $(1)@$(2) && \
		git sparse-checkout set *.proto && \
		git checkout tags/$(2) -b master; \
	}
endef

protoc-gen-go: ## Download protoc-gen-go locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/protoc-gen-go,google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1)
	$(call go-install,$(PROJECT_DIR)/bin/protoc-gen-go-grpc,google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0)
