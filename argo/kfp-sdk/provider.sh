#!/usr/bin/env ash

set -e -o pipefail

case $1 in
  "pipeline")
    NAME=$(yq e '.name' $3)
    VERSION=$(yq e '.version' $3)
    ENDPOINT=$(yq e '.endpoint' $4)
    case $2 in
      "create")
      kfp-ext --endpoint $ENDPOINT --output json pipeline upload --pipeline-name $NAME '/tmp/pipeline.yaml' | jq -r '."Pipeline Details"."Pipeline ID"'
      ;;
      "update")
      kfp-ext --endpoint $ENDPOINT --output json pipeline upload-version --pipeline-version $VERSION --pipeline-id $5 '/tmp/pipeline.yaml' | jq -r '."Version name"'
      ;;
      "delete")
      kfp-ext --endpoint $ENDPOINT --output json pipeline delete $5
      ;;
    esac
    ;;
  *)
    kfp-ext "$@"
    ;;
esac
