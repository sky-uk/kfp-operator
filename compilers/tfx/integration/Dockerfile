ARG UV_VERSION
ARG PYTHON_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:${PYTHON_VERSION}

COPY --from=uv /uv /bin/uv
COPY uv.lock pyproject.toml ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --locked --no-install-project --no-dev

# Patch TFX executor to enrich input artifact types from component inputs spec.
# KFP v2 driver does not populate instanceSchema on input artifacts resolved
# from upstream MLMD outputs, causing TFX to create generic Artifact objects.
COPY patch_tfx_executor.py /tmp/patch_tfx_executor.py
RUN python /tmp/patch_tfx_executor.py && rm /tmp/patch_tfx_executor.py

WORKDIR /pipeline
COPY *.py ./
COPY *.yaml ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
