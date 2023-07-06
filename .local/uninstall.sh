#!/usr/bin/env bash

set -eo pipefail

dir=$(realpath "$(dirname "$0")/..")

cd "$dir"
for r in "$dir"/.local/resources/*; do
  WEBHOOK_HOSTNAME=dummy WEBHOOK_CA_CERT=dummy envsubst < "$r" | kubectl delete -f -
done
