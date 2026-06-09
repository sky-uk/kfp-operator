ARG UV_VERSION
ARG TFX_PATCH_IMAGE=kfp-operator-tfx-patch
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv
FROM ${TFX_PATCH_IMAGE} AS tfx-patch

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

# Patch TFX for OSS KFP v2 compatibility (schema mapping, null schemas,
# artifact class inference, input enrichment, model dir flattening).
COPY --from=tfx-patch /patch_tfx_kfp_v2.py /tmp/patch_tfx_kfp_v2.py
RUN python /tmp/patch_tfx_kfp_v2.py && rm /tmp/patch_tfx_kfp_v2.py