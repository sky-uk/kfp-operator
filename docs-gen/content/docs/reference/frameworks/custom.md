---
title: "Adding a custom pipeline framework"
linkTitle: "Custom"
type: docs
weight: 3
---

Firstly a Docker image needs to be created that contains the necessary dependencies for the custom pipeline framework. 
This image should be pushed to a container registry that the KFP Operator deployment has access to.
e.g. `ghcr.io/kfp-operator/kfp-operator-custom-compiler:version-tag`

The Docker image needs to conform to the correct structure for the KFP Operator to be able to use it. See [readme](https://github.com/sky-uk/kfp-operator/blob/master/compilers/README.md) for more information.

Once the docker image is correctly published to an accessible repository it needs to be configured as an available frameworks for the provider it is to be used in.
This is done by creating or adding to the `frameworks` element in the [provider custom resource](../resources/provider/#common-fields).

Then to use the custom framework in a pipeline simply configure the framework attribute in the [pipeline resource](../resources/pipeline/#fields).

## Compiler Workflow
The `kfp-operator-create-compiled` workflow `compile` step accepts the following parameters:
- **resource-image**: the image containing the model code (looked up from the pipeline resource in preview workflow step)
- **pipeline-framework-image**: the image for the pipeline framework compiler

The image specified in `pipeline-framework-image` is executed as a initContainer and runs the entrypoint script. The entrypoint
script should copy the required compiler code into the shared directory `/shared` (which is mirrored into the main container) and then exit. This `/shared` location is passed as 
the first and only parameter to the entrypoint script.

Once the init container has complete then the main container is executed. The `/shared/compile.sh` which needs to be provided
by the compiler image should simply execute the compiler module.

See examples of entrypoint and compile scripts [here](https://github.com/sky-uk/kfp-operator/blob/master/compilers/resources).
