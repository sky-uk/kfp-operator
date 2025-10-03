FROM python:3.9.21-slim AS python39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*.whl && rm /*.whl
