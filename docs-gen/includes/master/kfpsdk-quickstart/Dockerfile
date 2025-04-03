FROM python:3.9.21-slim AS python39

COPY dist/kfpsdk_penguin_pipeline-*.whl .
RUN pip --no-cache-dir install /*.whl && rm /*.whl