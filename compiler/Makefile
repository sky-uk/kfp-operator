all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	poetry install

docker-build: build
	docker build . -t compiler
