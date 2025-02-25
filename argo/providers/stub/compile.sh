#!/bin/sh

while true; do
  case "$1" in
    --output_file)
      echo "{\"foo\": \"bar\"}" > "$2"
      break;;
    *)
      shift 2;;
  esac
done
