ARG UV_VERSION
ARG PYTHON_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:${PYTHON_VERSION}

COPY --from=uv /uv /bin/uv
COPY uv.lock pyproject.toml ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --locked --no-install-project --no-dev

WORKDIR /pipeline
COPY *.py ./
COPY *.yaml ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
