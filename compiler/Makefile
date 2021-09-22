all: build

##@ build

build:
	poetry install

docker:
	docker build . -t compiler

test: build
	poetry run pytest

##@ CI

prBuild: test docker

cdBuild: prBuild
