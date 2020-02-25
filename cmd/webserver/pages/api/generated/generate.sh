#!/usr/bin/env bash

curl \
  --silent \
  --output openapi.json \
  http://localhost:"$STEAM_PORT"/api/openapi.json

oapi-codegen \
  -o generated.go \
  -generate types,chi-server,spec \
  -package generated \
  openapi.json
