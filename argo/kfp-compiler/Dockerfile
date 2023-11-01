FROM python:3.9.13 AS PYTHON39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/kfp-compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM python:3.10.12 AS PYTHON310
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/kfp-compiler && \
    rm /*-${WHEEL_VERSION}-*.whl

FROM alpine:3.14.2

WORKDIR /
COPY --from=PYTHON39 /kfp-compiler kfp-compiler/py3.9
COPY --from=PYTHON310 /kfp-compiler kfp-compiler/py3.10
ADD resources/compile.sh kfp-compiler/compile.sh

ADD resources/entrypoint.sh entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
