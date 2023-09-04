# TFX build
FROM tensorflow/tfx:1.2.0

WORKDIR /pipeline
COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
