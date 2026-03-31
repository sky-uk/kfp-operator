ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.10.12
ARG WHEEL_VERSION
COPY --from=uv /uv /bin/uv

COPY pyproject.toml uv.lock ./
RUN uv sync --no-install-project --no-dev
ENV PATH="/.venv/bin:$PATH"
RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --no-deps /*.whl && rm /*.whl
