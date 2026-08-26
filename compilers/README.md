# Pipeline Framework Compilers
This directory contains the compiler code for each supported pipeline framework. To add support for a new pipeline framework, follow the instructions below:
The following should be added to all Dockerfiles:

```Dockerfile
COPY --from=base-common resources/compile.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
```

The compile argo workflow will always execute `/compile.sh` passing in the following arguments:

* `--pipeline_config` is the path to the yaml file that contains the pipeline resource definition.
* `--output_file` is the path to output the compiled pipeline definition to.

The Python module is required to be named `compiler`. The entry point to the Python compiler module should accept (or ignore these) parameters.

Suggestion would be to match the interface defined in the TFX compiler:

```python
    @click.command()
    @click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
    @click.option('--output_file', help='Output file path', required=True)
    def compile(pipeline_config: str, output_file: str):
```

The pipeline definition should be written to the output file defined in `--output_file`.

## Integration test matrix

`make integration-test` injects the built compiler into per-version model images and
compiles a sample pipeline in each.

### TFX (`compilers/tfx`)

The TFX compiler ships one ABI-specific build per Python version (`/compiler/py3.9`
through `/compiler/py3.12`). Each model image runs the latest supported TFX for its
Python version.

| Python | TFX | Env dir | Model image |
|--------|--------|-----------------------------------|-------------------|
| 3.9    | 1.9.1  | `integration/penguin-39`      | `penguin:3.9`     |
| 3.10   | 1.21.0 | `integration/penguin-310-121` | `penguin:3.10-121`|
| 3.11   | 1.21.0 | `integration/penguin-311-121` | `penguin:3.11-121`|
| 3.12   | 1.21.0 | `integration/penguin-312-121` | `penguin:3.12-121`|

### KFP-SDK (`compilers/kfp-sdk`)

The KFP-SDK compiler is a single pure-Python build (built on 3.12, pinning `click<8.2` so
it still imports on 3.9). The matrix covers the oldest and newest supported Python
versions.

| Python | KFP           | Env dir / source                        | Model image                      |
|--------|---------------|-----------------------------------------|----------------------------------|
| 3.9    | >=2.17.0,<3   | `integration/quickstart-39`             | `kfpsdk-quickstart:3.9`          |
| 3.12   | >=2.17.0,<3   | `docs-gen/includes/master/kfpsdk-quickstart` | `kfp-operator-kfpsdk-quickstart` |
