ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.10.12
ARG WHEEL_VERSION

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        libbrotli1=1.0.9-2+deb12u1 \
        libbrotli-dev=1.0.9-2+deb12u1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=uv /uv /bin/uv

COPY pyproject.toml uv.lock ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --no-install-project --no-dev

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --system --no-deps /*.whl && rm /*.whl
