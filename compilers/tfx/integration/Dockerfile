ARG PYTHON_VERSION

FROM python:$PYTHON_VERSION

ARG TFX_VERSION

RUN pip install tfx==$TFX_VERSION
RUN pip install -U google-cloud-aiplatform "shapely<2"

WORKDIR /pipeline
COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
