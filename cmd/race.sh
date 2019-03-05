#!/usr/bin/env bash

cd ${STEAM_PATH}
GOFLAGS=-mod=vendor go run -race *.go
