#!/usr/bin/env bash

files=(
  "google.golang.org/protobuf/cmd/protoc-gen-go"
  "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
  "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
  "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
)

for i in "${files[@]}"; do
  if [ ! -f $(which $(basename $i)) ]; then
    go get -u "$i"
  fi
done

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
  "$STEAM_PATH"/pkg/backend/protos/*.proto

#  --openapiv2_out "$STEAM_PATH"/cmd/api/generated_openapi \
#  --openapiv2_opt logtostderr=true \

echo "Done"
