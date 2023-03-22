include ../../../common.mk

IMG := kfp-operator-stub-provider

##@ Build

build:
	go build -o bin/provider ./cmd

##@ Containers

include ../../../docker-targets.mk
