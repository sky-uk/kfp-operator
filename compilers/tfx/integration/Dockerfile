ARG UV_VERSION
ARG PYTHON_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:${PYTHON_VERSION}

COPY --from=uv /uv /bin/uv
COPY uv.lock pyproject.toml /deps/
RUN cd /deps && uv export --locked --no-dev --no-emit-project > requirements.txt && \
    uv pip install --system -r requirements.txt

WORKDIR /pipeline
COPY *.py ./
COPY *.yaml ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
