#!/bin/sh

export PYTHONPATH="$PYTHONPATH:$(dirname $0)"

python3 -m compiler "$@"
