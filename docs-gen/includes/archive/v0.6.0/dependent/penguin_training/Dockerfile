# TFX build
FROM tensorflow/tfx:1.13.0

WORKDIR /pipeline
COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
