---
title: "Debugging"
weight: 5
---

## Kubernetes Events

The operator emits Kubernetes events for all resource transitions which can be viewed using `kubectl describe`.

Example:

```shell 
$ kubectl describe pipeline pipeline-sample
...
Events:
  Type     Reason      Age    From          Message
  ----     ------      ----   ----          -------
  Normal   Syncing     5m54s  kfp-operator  Updating [version: "v5-841641"]
  Warning  SyncFailed  101s   kfp-operator  Failed [version: "v5-841641"]: pipeline update failed
  Normal   Syncing     9m47s  kfp-operator  Updating [version: "57be7f4-681dd8"]
  Normal   Synced      78s    kfp-operator  Succeeded [version: "57be7f4-681dd8"]
```

## Compiling locally

The KFP-Operator's compiler can be used locally for debugging purposes. This can be especially useful for troubleshooting environment variable and beam argument resolution.

### Environment setup and compiler injection

The compiler is injected into a shared directory first before it can be called from within the pipeline image.
Note that the setup is usually only needed once unless you want to use a different version of the compiler.

```shell
export KFP_COMPILER_IMAGE=ghcr.io/kfp-operator/kfp-operator-tfx-compiler:<KFP-Operator version>
docker pull $KFP_COMPILER_IMAGE

# Create a temporary directory for the following steps, alternatively choose a different location
SHARED_DIR=$(mktemp -d)

# Inject the compiler into the shared temporary directory
docker run -v $SHARED_DIR:/shared $KFP_COMPILER_IMAGE /shared
```

### Compiler configuration

The compilation process can be configured as follows:

```shell
export PIPELINE_IMAGE=<your pipeline image>
# Choose an execution mode: v1 for KFP or v2 for Vertex AI
export EXECUTION_MODE=v1

# create the compiler configuration
cat > $SHARED_DIR/config.yaml << EOF
name: <Your pipeline name>
rootLocation: <pipeline root location. for debugging, this can be any string>
servingLocation: <model serving location. for debugging, this can be any string>
image: $PIPELINE_IMAGE
tfxComponents: <component function>
env:
  <Dict[str, str] of environment variables to be passed to the compilation step>
beamArgs:
  <Dict[str, List[str]] of beam arguments>
EOF
```

### Running the compiler

You can then run the compiler from inside your pipeline container to produce `$SHARED_DIR/pipeline_out.yaml`.

```shell
# Run the compiler in your pipeline image
docker run -v $SHARED_DIR:/shared --entrypoint /shared/compile.sh $PIPELINE_IMAGE --pipeline_config /shared/config.yaml --output_file /shared/pipeline_out.yaml --execution_mode $EXECUTION_MODE
```
