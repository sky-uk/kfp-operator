FROM python:3.10.12
ARG WHEEL_VERSION

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

RUN pip install poetry==1.7.1

COPY pyproject.toml .
COPY poetry.lock .

RUN poetry config virtualenvs.create false
RUN poetry install --no-root
WORKDIR /pipeline

COPY penguin_pipeline/*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
