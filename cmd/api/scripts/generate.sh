#!/usr/bin/env bash

oapi-codegen \
  -o ../generated/generated.go \
  -generate types,chi-server,spec \
  -package generated \
  http://localhost:"$STEAM_PORT"/api/gamedb.json
