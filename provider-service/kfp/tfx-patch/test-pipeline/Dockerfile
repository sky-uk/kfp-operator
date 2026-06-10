ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.10.12

COPY --from=uv /uv /bin/uv

COPY pyproject.toml uv.lock ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --locked --no-install-project --no-dev

# Install the shim (import hook approach — no source file modification)
COPY tfx_kfp_v2_shim/ /tmp/tfx_kfp_v2_shim/
RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

# Install pipeline code as a package into site-packages so it's importable
# from the compiler (which runs via compile.sh with its own PYTHONPATH)
COPY penguin_pipeline/ /usr/local/lib/python3.10/site-packages/penguin_pipeline/
RUN touch /usr/local/lib/python3.10/site-packages/penguin_pipeline/__init__.py
