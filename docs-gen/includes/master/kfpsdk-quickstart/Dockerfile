FROM python:3.13.2-slim AS python39
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*.whl && rm /*.whl
