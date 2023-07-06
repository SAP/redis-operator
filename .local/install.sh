#!/usr/bin/env bash

set -eo pipefail

WEBHOOK_HOSTNAME=$1
if [ -z "$WEBHOOK_HOSTNAME" ]; then
  WEBHOOK_HOSTNAME=host.internal
fi

dir=$(realpath "$(dirname "$0")/..")
cd "$dir"
mkdir -p tmp/ssl
cd tmp/ssl

cfssl gencert -initca "$dir"/.local/ssl/ca.json | cfssljson -bare ca
cfssl gencert -ca ca.pem -ca-key ca-key.pem "$dir"/.local/ssl/webhook.json  | cfssljson -bare webhook
rm -f *.csr tls.key tls.crt
ln -s webhook-key.pem tls.key
ln -s webhook.pem tls.crt

cd "$dir"
for r in "$dir"/.local/resources/*; do
  WEBHOOK_HOSTNAME="$WEBHOOK_HOSTNAME" WEBHOOK_CA_CERT=$(cat "$dir"/tmp/ssl/ca.pem | openssl base64 -e -A) envsubst < "$r" | kubectl apply -f -
done
