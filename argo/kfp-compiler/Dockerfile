FROM python:3.7.12 AS PYTHON

ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/kfp-compiler

FROM alpine:3.14.2

COPY --from=PYTHON /kfp-compiler /kfp-compiler
ADD resources/compile.py /kfp-compiler/compile.py

ADD resources/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
