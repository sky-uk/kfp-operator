# TFX build 
FROM tensorflow/tfx:1.2.0

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

WORKDIR /pipeline
COPY pipeline.py ./
COPY trainer.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
