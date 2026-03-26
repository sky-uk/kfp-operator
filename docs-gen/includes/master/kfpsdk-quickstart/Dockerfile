FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION
ARG UV_VERSION
COPY --from=ghcr.io/astral-sh/uv:${UV_VERSION} /uv /bin/uv

COPY pyproject.toml uv.lock ./
RUN uv sync --no-install-project --no-dev

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --no-deps /*.whl && rm /*.whl
