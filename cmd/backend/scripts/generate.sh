#!/usr/bin/env bash

protoc \
  --proto_path ../../pkg/backend/protos \
  --go_out=plugins=grpc:../../pkg/backend/generated \
  ../../pkg/backend/protos/*.proto
