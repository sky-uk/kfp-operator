ARG PYTHON_VERSION

FROM python:$PYTHON_VERSION

ARG TFX_VERSION

RUN pip install --prefer-binary --no-cache-dir tfx==$TFX_VERSION
RUN pip install --no-cache-dir --upgrade google-cloud-aiplatform "shapely<2"

WORKDIR /pipeline
COPY *.py ./
COPY *.yaml ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"

