# Pipeline Framework Compilers

The following should be added to all Dockerfiles:
```Dockerfile
COPY --from=base-common resources/compile.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
```

The compile argo workflow will always execute /compile.sh passing in the following arguments, `--pipeline_config`, `--provider_config`, and `--output_file`.

* `--pipeline_config` is a path to a yaml file that contains the pipeline resource definition.
* `--provider_config` is a path to a yaml file that contains the provider configuration.
* `--output_file` is a path to a file that the compiled pipeline definition should be written to.

The Python module is required to be named `compiler`. The entry point to the Python compiler module should accept (or ignore these) parameters.

Suggestion would be to match the interface defined in the TFX compiler:

```python
    @click.command()
    @click.option('--pipeline_config', help='Pipeline configuration in yaml format', required=True)
    @click.option('--provider_config', help='Provider configuration in yaml format', required=True)
    @click.option('--output_file', help='Output file path', required=True)
    def compile(pipeline_config: str, provider_config: str, output_file: str):
```

The pipeline definition should be written to the output file defined in `--output_file`.
