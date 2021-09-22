all: build

##@ Development

test: build
	pip install poetry
	poetry run pytest

##@ Build

build:
	poetry install

docker-build: build
	docker build . -t compiler
