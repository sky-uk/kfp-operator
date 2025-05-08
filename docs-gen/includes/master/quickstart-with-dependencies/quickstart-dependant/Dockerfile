FROM python:3.10.12
ARG WHEEL_VERSION

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*.whl && rm /*.whl
