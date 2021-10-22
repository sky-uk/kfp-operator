include ../../version.mk

IMG := kfp-operator-argo-compiler

all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	poetry install
	@echo Nothing to build.

##@ Containers

include ../../docker-targets.mk
