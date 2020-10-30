#!/usr/bin/env bash

protoc \
  --go_out=../../pkg/backend/generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=../../pkg/backend/generated \
  --go-grpc_opt=paths=source_relative \
  --proto_path ../../pkg/backend/protos/ \
  ../../pkg/backend/protos/*.proto
