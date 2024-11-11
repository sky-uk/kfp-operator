FROM nvidia/cuda:11.8.0-runtime-ubuntu22.04

# RUN apt-get update
# RUN apt-get upgrade -y

# Install dependencies for pyenv
RUN apt-get update && \
    apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libffi-dev \
    liblzma-dev && \    
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

# Install pyenv
RUN curl https://pyenv.run | bash

# ENV PYENV_ROOT="/root/.pyenv"
# ENV PATH="$PYENV_ROOT/bin:$PATH"

# RUN eval "$(pyenv init -)"

# # Set up environment variables for pyenv
ENV PATH="/root/.pyenv/bin:$PATH"
ENV PYENV_ROOT="/root/.pyenv"
ENV PATH="$PYENV_ROOT/shims:$PYENV_ROOT/bin:$PATH"

# Initialize pyenv
RUN echo 'eval "$(pyenv init --path)"' >> ~/.bashrc

# Install a specific Python version (e.g., 3.10.15)
RUN pyenv install 3.10.12 && pyenv global 3.10.12

# Verify python installation
RUN python --version

# Update pip
RUN pip install --upgrade pip

# Verify pip installation
RUN pip --version

# Install dependencies for nvidia python
RUN apt-get update && \
    apt-get install -y \
    libnvinfer8 && \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

# apt-get install libcudnn8

# Install a specific version of Poetry
# RUN curl -sSL https://install.python-poetry.org | POETRY_VERSION=1.8.2 python -

# Add Poetry to PATH
# ENV PATH="/root/.local/bin:$PATH"

# Disable Poetry's virtual environment creation
# ENV POETRY_VIRTUALENVS_CREATE=false

# Copy the poetry files
# COPY pyproject.toml poetry.lock* ./
# COPY pyproject.toml ./

# Install dependencies from cache if available
# RUN poetry install --no-root

# -------------------------------------------------------------------------
# Install beam dataflow workers
# https://cloud.google.com/dataflow/docs/guides/using-custom-containers
# Copy files from official SDK image, including script/dependencies.
COPY --from=apache/beam_python3.10_sdk:2.47.0 /opt/apache/beam /opt/apache/beam
# Set the entrypoint to Apache Beam SDK launcher.
ENTRYPOINT ["/opt/apache/beam/boot"]

# ------------------------------------------------------
#         Example usage of the base docker
# ------------------------------------------------------

# Copy app code
# COPY app /app

# Install dependencies from cache if available
# RUN poetry install --no-root --directory /app

# WORKDIR /app






# FROM python:3.10.12 as builder

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

RUN pip install poetry==1.7.1

COPY pyproject.toml .
COPY poetry.lock .

RUN poetry config virtualenvs.create true
RUN poetry config virtualenvs.in-project true
RUN poetry install --no-root

# FROM python:3.10.12-slim as runtime

# ENV VIRTUAL_ENV=/.venv \
#     PATH="/.venv/bin:$PATH"

WORKDIR /pipeline

# COPY --from=builder ${VIRTUAL_ENV} ${VIRTUAL_ENV}
# COPY --from=builder /data /data

COPY penguin_pipeline/*.py ./

# ENV PYTHONPATH="/pipeline:${PYTHONPATH}"
