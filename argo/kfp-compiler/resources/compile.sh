#!/bin/sh

PYTHON_VERSION=$(python3 -V | grep -Eo '[0-9]+\.[0-9]+')

export PYTHONPATH="$PYTHONPATH:$(dirname $0)/py$PYTHON_VERSION"

python3 -m 'kfp_compiler' "$@"
