FROM python:3.10.5-alpine3.16

ARG WHEEL_VERSION

RUN apk add jq yq

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl && \
    rm /*-${WHEEL_VERSION}-*.whl
COPY provider.sh /

ENTRYPOINT ["/provider.sh"]
