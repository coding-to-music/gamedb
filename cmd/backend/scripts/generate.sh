#!/usr/bin/env bash

protoc \
  -I "$STEAM_PATH"/pkg/backend/protos \
  --go_out="$STEAM_PATH"/pkg/backend/generated \
  --go_opt=paths=source_relative \
  --go-grpc_out="$STEAM_PATH"/pkg/backend/generated \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out "$STEAM_PATH"/pkg/backend/generated \
  --grpc-gateway_opt logtostderr=true \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt generate_unbound_methods=true \
  --openapiv2_out "$STEAM_PATH"/cmd/api/generated_openapi \
  --openapiv2_opt logtostderr=true \
  "$STEAM_PATH"/pkg/backend/protos/*.proto

echo "Done"
