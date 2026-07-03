ARG UV_VERSION
FROM ghcr.io/astral-sh/uv:${UV_VERSION} AS uv

FROM python:3.10.19-trixie
ARG WHEEL_VERSION
COPY --from=uv /uv /bin/uv

# Force pure-Python protobuf to avoid C++ ABI conflicts between
# protobuf 4.x and grpcio's bundled C++ protobuf library.
ENV PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python

COPY pyproject.toml uv.lock ./
ENV UV_PROJECT_ENVIRONMENT=/usr/local
RUN uv sync --locked --no-install-project --no-dev --python 3.10

# Install the tfx_kfp_v2_shim: runtime TFX patches for OSS KFP v2 compatibility
# (artifact-type resolution) plus the executor force-exit. Sourced via the
# tfx-shim named build context so the image is self-contained.
COPY --from=tfx-shim . /tmp/tfx_kfp_v2_shim/
RUN python /tmp/tfx_kfp_v2_shim/install_shim.py && rm -rf /tmp/tfx_kfp_v2_shim

RUN mkdir /data
RUN wget https://raw.githubusercontent.com/tensorflow/tfx/master/tfx/examples/penguin/data/labelled/penguins_processed.csv -P /data

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN uv pip install --system --no-deps /*.whl && rm /*.whl
