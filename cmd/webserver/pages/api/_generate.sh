#!/usr/bin/env bash

curl \
  --output _gamedb.json \
  http://localhost:"$STEAM_PORT"/api/gamedb.json

oapi-codegen \
  -o ./generated/generated.go \
  -generate types,chi-server,spec \
  -package generated \
  _gamedb.json
