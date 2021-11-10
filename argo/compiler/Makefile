include ../../version.mk
include ../../newline.mk

IMG := kfp-operator-argo-compiler

all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	pip install poetry-dynamic-versioning --quiet
	poetry install
	@echo Nothing to build.

##@ Containers

include ../../docker-targets.mk
