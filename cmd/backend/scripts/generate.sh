#!/usr/bin/env bash

BASE=/Users/james/Websites/gamedb/gamedb

protoc \
  -I $BASE/pkg/backend/protos \
  --go_out=$BASE/pkg/backend/generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=$BASE/pkg/backend/generated \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out $BASE/pkg/backend/generated \
  --grpc-gateway_opt logtostderr=true \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt generate_unbound_methods=true \
  --openapiv2_out $BASE/cmd/api/generated_openapi \
  --openapiv2_opt logtostderr=true \
  $BASE/pkg/backend/protos/*.proto
