FROM python:3.7.12

ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl

ENTRYPOINT ["mlmd"]
