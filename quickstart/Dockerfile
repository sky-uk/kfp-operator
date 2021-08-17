FROM tensorflow/tfx:1.0.0

WORKDIR /pipeline
COPY penguin_pipeline/*.py ./
ENV PYTHONPATH="/pipeline:${PYTHONPATH}"