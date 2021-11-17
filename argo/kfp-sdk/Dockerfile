FROM python:3.7.12-alpine3.14

ARG WHEEL_VERSION

RUN apk add jq

COPY dist /
RUN pip install /*-${WHEEL_VERSION}-*.whl

ENTRYPOINT ["kfp-ext"]
