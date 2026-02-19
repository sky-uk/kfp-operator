ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION

COPY --from=uv /uv /bin/uv
COPY uv.lock pyproject.toml /
COPY dist/*-${WHEEL_VERSION}-*.whl /

ENV UV_PROJECT_ENVIRONMENT=/compiler
RUN uv sync --locked --no-install-project --no-dev --python 3.9

RUN uv pip install --no-deps /*-${WHEEL_VERSION}-*.whl --target=/compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM alpine:3.22.0

WORKDIR /

COPY --from=python39 /compiler compiler/

COPY --from=base-common resources/compile-any.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
