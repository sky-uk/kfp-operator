#!/bin/ash

echo "BADGER"
echo $1
echo $2
echo "BADGER"

if [ "$#" -ne 2 ] || ! [ -d "$2" ]; then
  echo "Usage: $(basename $0) <destination directory>" >&2
  exit 1
fi

cp -r /$1-compiler/* $2

echo "Done injecting compiler"
