FROM python:3.7.12 AS PYTHON

ARG WHEEL_VERSION

# change this in other dockerfiles as well
COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl --target=/kfp-compiler

FROM alpine:3.14.2

COPY --from=PYTHON /kfp-compiler /kfp-compiler
ADD compile.py /kfp-compiler/compile.py

ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
