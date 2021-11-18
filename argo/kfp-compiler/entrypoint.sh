#!/bin/ash

if [ "$#" -ne 1 ] || ! [ -d "$1" ]; then
  echo "Usage: $(basename $0) <destination directory>" >&2
  exit 1
fi

cp -r /kfp-compiler/* $1

echo "Done injecting compiler"
