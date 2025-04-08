FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION

COPY dist/kfpsdk_quickstart-${WHEEL_VERSION}-*.whl .
RUN pip --no-cache-dir install /*.whl && rm /*.whl
