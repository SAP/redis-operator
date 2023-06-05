#!/usr/bin/env bash

set -eo pipefail

CHART_VERSION=$1
if [ -z "$CHART_VERSION" ]; then
  echo "Warning: no chart version specified; updating to latest chart version"
fi

cd $(dirname "${BASH_SOURCE[0]}")/..

mkdir -p pkg/operator/data/charts
pushd pkg/operator/data/charts >/dev/null
rm -rf redis redis-*.tgz
if [ -z "$CHART_VERSION" ]; then
  helm pull oci://registry-1.docker.io/bitnamicharts/redis
else
  helm pull oci://registry-1.docker.io/bitnamicharts/redis --version $CHART_VERSION
fi
tar xfz redis-*.tgz
rm -f redis-*.tgz
cd redis
helm dep build
rm -f charts/*.tgz
popd >/dev/null