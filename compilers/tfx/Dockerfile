FROM python:3.9.13 AS PYTHON39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/tfx-compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM python:3.10.12 AS PYTHON310
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/tfx-compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM alpine:3.20.1

WORKDIR /
COPY --from=PYTHON39 /tfx-compiler tfx-compiler/py3.9
COPY --from=PYTHON310 /tfx-compiler tfx-compiler/py3.10

ADD resources/compile.sh tfx-compiler/compile.sh
ADD resources/entrypoint.sh entrypoint.sh

USER 65534:65534
ENTRYPOINT ["/entrypoint.sh"]
