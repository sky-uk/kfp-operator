FROM python:3.10.12
ARG WHEEL_VERSION

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*.whl && rm /*.whl
