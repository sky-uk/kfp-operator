#!/bin/sh

export PYTHONPATH="$PYTHONPATH:$(dirname $0)"

export KFP_DISABLE_EXECUTION_CACHING_BY_DEFAULT=true

python3 -m compiler "$@"
