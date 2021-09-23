all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	pyenv install -s
	command -v poetry || pip install poetry
	poetry install

docker-build: build
	docker build . -t compiler
