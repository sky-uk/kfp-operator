# TFX build 
FROM tensorflow/tfx:1.2.0

RUN pip install tfx==1.2.0
RUN pip install kfp
RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

WORKDIR /pipeline
COPY penguin_pipeline/*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
