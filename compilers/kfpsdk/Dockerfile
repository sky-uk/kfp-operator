FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM alpine:3.20.1

WORKDIR /

COPY --from=python39 /compiler compiler/

COPY --from=base-common resources/compile-any.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
