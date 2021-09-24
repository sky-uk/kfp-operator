all: build

##@ Development

test: build
	poetry install
	poetry run pytest

##@ Build

build:
	@echo Nothing to build.

docker-build: build
	docker build . -t compiler
