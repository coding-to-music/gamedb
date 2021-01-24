#!/usr/bin/env bash

files=(
  "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
)

for i in "${files[@]}"; do
  go get -u "$i"
done

oapi-codegen \
  -o ./generated/generated.go \
  -generate types,chi-server,spec \
  -package generated \
  http://localhost:"$STEAM_PORT"/api/gamedb.json

echo "Done"
