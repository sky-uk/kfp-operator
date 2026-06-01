ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.10-trixie
ARG WHEEL_VERSION
COPY --from=uv /uv /bin/uv

COPY pyproject.toml uv.lock ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --locked --no-install-project --no-dev --python 3.10

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --system --no-deps /*.whl && rm /*.whl

# Patch TFX for KFP v2 compatibility:
# 1. compiler_utils: Map system.* and tfx.* schema titles to TFX class paths.
#    KFP v2 backend passes artifacts with system.* titles (e.g. system.Model)
#    instead of tfx.* titles. Without this mapping, TFX tries to import
#    "system.Model" as a Python path, causing ImportError.
# 2. kubeflow_v2_entrypoint_utils:
#    a. Handle None instance_schema gracefully (yaml.safe_load returns None).
#    b. Replace _parse_raw_artifact to infer the correct TFX artifact class
#       from metadata keys when KFP v2 uses ambiguous system.* schema titles.
RUN python3 -c "\
from pathlib import Path; \
import tfx.orchestration.kubeflow.v2.compiler_utils as m; \
p = Path(m.__file__); \
p.write_text(p.read_text() + '''\n\
# KFP v2 system.* schema title compatibility patch\n\
TITLE_TO_CLASS_PATH.update({\n\
    \"system.Artifact\": name_utils.get_full_name(simple_artifacts.File),\n\
    \"system.Model\": name_utils.get_full_name(standard_artifacts.Model),\n\
    \"system.Dataset\": name_utils.get_full_name(simple_artifacts.Dataset),\n\
    \"system.Metrics\": name_utils.get_full_name(simple_artifacts.Metrics),\n\
    \"system.Statistics\": name_utils.get_full_name(simple_artifacts.Statistics),\n\
    \"tfx.Artifact\": name_utils.get_full_name(simple_artifacts.File),\n\
    \"tfx.Model\": name_utils.get_full_name(standard_artifacts.Model),\n\
    \"tfx.Dataset\": name_utils.get_full_name(simple_artifacts.Dataset),\n\
    \"tfx.Metrics\": name_utils.get_full_name(simple_artifacts.Metrics),\n\
    \"tfx.Statistics\": name_utils.get_full_name(simple_artifacts.Statistics),\n\
})\n\
''')" && \
python3 -c "\
from pathlib import Path; \
import tfx.orchestration.kubeflow.v2.container.kubeflow_v2_entrypoint_utils as m; \
p = Path(m.__file__); \
src = p.read_text(); \
src = src.replace(\
    'data = yaml.safe_load(type_schema.instance_schema)', \
    'data = yaml.safe_load(type_schema.instance_schema) or {}'); \
p.write_text(src)" && \
python3 << 'PYEOF'
# Patch _parse_raw_artifact to infer the correct TFX artifact class from
# the artifact metadata when the schema title is a generic system.* type.
# KFP v2 maps multiple TFX types to a single system type (e.g. both
# tfx.Examples and tfx.Dataset become system.Dataset). We inspect the
# artifact metadata keys to pick the right TFX class.
from pathlib import Path
import tfx.orchestration.kubeflow.v2.container.kubeflow_v2_entrypoint_utils as m

patch = '''

# --- Begin KFP v2 artifact class inference patch ---
_SYSTEM_DATASET_PROPERTY_MAP = {
    frozenset({"split_names", "span", "version"}): "tfx.Examples",
    frozenset({"split_names", "span"}): "tfx.Examples",
    frozenset({"split_names"}): "tfx.Examples",
}

_original_parse_raw_artifact = _parse_raw_artifact

def _patched_parse_raw_artifact(artifact_pb, name_from_id):
    import logging as _logging
    _type_schema = artifact_pb.type
    _title = None
    if _type_schema.WhichOneof("kind") == "schema_title":
        _title = _type_schema.schema_title
    elif _type_schema.WhichOneof("kind") == "instance_schema":
        _data = yaml.safe_load(_type_schema.instance_schema) or {}
        _title = _data.get("title")
    if _title in ("system.Dataset", "tfx.Dataset"):
        from google.protobuf import json_format as _jf
        _meta_keys = set()
        if artifact_pb.metadata and artifact_pb.metadata.fields:
            _meta_keys = set(artifact_pb.metadata.fields.keys())
        _prop_keys = set(artifact_pb.properties.keys()) if artifact_pb.properties else set()
        _all_keys = _meta_keys | _prop_keys
        for _sig, _cls_title in _SYSTEM_DATASET_PROPERTY_MAP.items():
            if _sig <= _all_keys:
                _logging.info("Inferred TFX type %s from metadata keys %s", _cls_title, _all_keys)
                if _type_schema.WhichOneof("kind") == "schema_title":
                    artifact_pb.type.schema_title = _cls_title
                else:
                    _data["title"] = _cls_title
                    artifact_pb.type.instance_schema = yaml.dump(_data, sort_keys=False)
                break
    return _original_parse_raw_artifact(artifact_pb, name_from_id)

_parse_raw_artifact = _patched_parse_raw_artifact
# --- End KFP v2 artifact class inference patch ---
'''

p = Path(m.__file__)
p.write_text(p.read_text() + patch)
PYEOF

# 3. kubeflow_v2_run_executor: Enrich input artifact types from the component
#    inputs spec before parsing. KFP v2 driver passes empty instanceSchema for
#    input artifacts resolved from upstream MLMD outputs. This copies the correct
#    artifact type from the inputs_spec to each input artifact.
COPY patch_tfx_executor.py /tmp/patch_tfx_executor.py
RUN python3 /tmp/patch_tfx_executor.py && rm /tmp/patch_tfx_executor.py

# 4. path_utils: Remove the Format-Serving subdirectory from serving_model_dir.
#    TFX saves trained models under model_uri/Format-Serving/ by default. KFP v2
#    launcher has a bug downloading deeply nested directory artifacts from GCS
#    (mkdir fails when a file placeholder exists for a directory name). Saving
#    directly to model_uri avoids this issue.
RUN python3 -c "\
from pathlib import Path; \
import tfx.utils.path_utils as m; \
p = Path(m.__file__); \
src = p.read_text(); \
old = 'return os.path.join(output_uri, path_constants.SERVING_MODEL_DIR)'; \
new = 'return output_uri  # patched: skip Format-Serving subdir for KFP v2 compat'; \
assert old in src, f'Could not find target string in {p}'; \
src = src.replace(old, new); \
p.write_text(src); \
print('Patched path_utils.serving_model_dir')"