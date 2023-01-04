#!/bin/sh

while true; do
  case "$1" in
    --output_file)
      echo "resource" > "$2"
      break;;
    *)
      shift 2;;
  esac
done
