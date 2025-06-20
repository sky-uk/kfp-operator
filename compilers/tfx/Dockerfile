FROM python:3.9.13 AS python39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM python:3.10.12 AS python310
ARG WHEEL_VERSION

COPY dist/compiler-${WHEEL_VERSION}-*.whl /
RUN pip install /compiler-${WHEEL_VERSION}-*.whl --target=/compiler
RUN rm /compiler-${WHEEL_VERSION}-*.whl

FROM alpine:3.22.0

WORKDIR /
COPY --from=python39 /compiler compiler/py3.9
COPY --from=python310 /compiler compiler/py3.10

COPY --from=base-common resources/compile.sh compiler/compile.sh
COPY --from=base-common resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
