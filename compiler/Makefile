all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	command -v poetry || pip install poetry
	poetry install

docker-build: build
	docker build . -t compiler
