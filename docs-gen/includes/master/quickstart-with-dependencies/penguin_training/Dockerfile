FROM python:3.10.12 as builder

RUN pip install poetry==1.7.1

COPY pyproject.toml .
COPY poetry.lock .

RUN poetry config virtualenvs.create true
RUN poetry config virtualenvs.in-project true
RUN poetry install --no-root

FROM python:3.10.12-slim as runtime

ENV VIRTUAL_ENV=/.venv \
    PATH="/.venv/bin:$PATH"

WORKDIR /pipeline

COPY --from=builder ${VIRTUAL_ENV} ${VIRTUAL_ENV}

COPY ./*.py ./

ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
