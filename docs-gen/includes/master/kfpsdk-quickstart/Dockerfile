ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.12.14-slim AS python312
ARG WHEEL_VERSION
COPY --from=uv /uv /bin/uv

COPY pyproject.toml uv.lock ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --no-install-project --no-dev

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --system --no-deps /*.whl && rm /*.whl
