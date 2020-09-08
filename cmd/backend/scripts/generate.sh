#!/usr/bin/env bash

protoc \
  --go_opt=paths=source_relative \
  --proto_path ../../pkg/backend/protos \
  --go_out=plugins=grpc:../../pkg/backend/generated \
  ../../pkg/backend/protos/*.proto
