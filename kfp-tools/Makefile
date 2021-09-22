all: docker

##@ build

docker:
	docker build . -t compiler

##@ CI

prBuild: docker

cdBuild: prBuild
