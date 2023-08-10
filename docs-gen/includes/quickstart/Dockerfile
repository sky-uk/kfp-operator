# TFX build
FROM tensorflow/tfx:1.2.0
ARG PIPELINE_DIR

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

WORKDIR /pipeline
COPY ${PIPELINE_DIR}/*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
