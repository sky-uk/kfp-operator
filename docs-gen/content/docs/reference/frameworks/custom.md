---
title: "Adding a custom pipeline framework"
linkTitle: "Custom"
type: docs
weight: 3
---

If your desired framework is not [natively supported by the KFP Operator](../), you can provide a custom Docker image that contains the necessary dependencies and code to compile pipelines using your desired framework.

This image should be pushed to a container registry that the KFP Operator deployment has access to. e.g. `ghcr.io/kfp-operator/kfp-operator-custom-compiler:version-tag`

Follow these steps to build the image and configure your installation of the KFP Operator:
1. Follow the steps in the [compilers README](https://github.com/sky-uk/kfp-operator/blob/master/compilers/README.md) to build your custom Docker image, ensuring it conforms to the correct structure. Examples of the structure can be found in the code for the natively supported frameworks. This image will be called by a set of Argo Workflows, with [these parameters](#compiler-workflow).
2. Publish the Docker image to a repository accessible via the KFP Operator deployment. 
3. Update your [Provider](../providers/overview/) resource to support your custom framework by specifying your framework name and image in `spec.frameworks[]`.
4. To then use the custom framework in a [Pipeline](../resources/pipeline/#fields) resource, simply configure `spec.framework` to take the same name as the framework set in the Provider resource, and any additional parameters that your framework requires.

### Compiler Workflow
The `kfp-operator-create-compiled` workflow `compile` step accepts the following parameters:
- **resource-image**: the image containing the model code (looked up from the pipeline resource in preview workflow step)
- **pipeline-framework-image**: the image for the pipeline framework compiler

The image specified in `pipeline-framework-image` is executed as a initContainer and runs the entrypoint script. The entrypoint
script should copy the required compiler code into the shared directory `/shared` (which is mirrored into the main container) and then exit. This `/shared` location is passed as 
the first and only parameter to the entrypoint script.

Once the init container has complete then the main container is executed. The `/shared/compile.sh` which needs to be provided
by the compiler image should simply execute the compiler module.

See examples of entrypoint and compile scripts [here](https://github.com/sky-uk/kfp-operator/blob/master/compilers/resources).
