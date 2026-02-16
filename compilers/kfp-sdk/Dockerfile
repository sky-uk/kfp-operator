FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN --mount=type=cache,target=/root/.cache/pip  \
    pip install /*-${WHEEL_VERSION}-*.whl --target=/compiler &&  \
    rm /*-${WHEEL_VERSION}-*.whl

FROM alpine:3.22.0

WORKDIR /

COPY --from=python39 /compiler compiler/

COPY --from=base-common resources/compile-any.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
