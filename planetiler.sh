#!/usr/bin/env bash

docker run --rm \
  -e JAVA_TOOL_OPTIONS="-Xmx2g" \
  -v "$(pwd)/data":/data \
  ghcr.io/onthegomap/planetiler:latest \
  /data/planetiler.yaml \
  --download \
  --output=/data/great-britain.mbtiles \
  --force