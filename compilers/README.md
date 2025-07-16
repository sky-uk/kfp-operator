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
