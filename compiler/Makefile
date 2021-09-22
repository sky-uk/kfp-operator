all: build

##@ build

build:
	poetry install

docker-build:
	docker build . -t compiler

test: build
	poetry run pytest
