#!/usr/bin/env bash

mkdir -p secret/base64

ssh-keygen -t ed25519 -f secret/ed25519 -N "" -C ""

if [[ "$(uname)" == "Darwin" ]]; then
    cat secret/ed25519 | base64 -b 0 >> secret/base64/ed25519
    cat secret/ed25519.pub | base64 -b 0 >> secret/base64/ed25519.pub
else
    cat secret/ed25519 | base64 -w 0 >> secret/base64/ed25519
    cat secret/ed25519.pub | base64 -w 0 >> secret/base64/ed25519.pub
fi