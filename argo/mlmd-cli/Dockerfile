FROM python:3.7.12

ARG WHEEL_VERSION

COPY dist /
RUN pip install /*${WHEEL_VERSION}*.whl

ENTRYPOINT ["mlm-cli"]
