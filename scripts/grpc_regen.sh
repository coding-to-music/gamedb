#!/usr/bin/env bash

# Settings
OUT=${STEAM_INFRASTRUCTURE_PATH}/grpc
COUNTRY="GBR"
BRAND="Game DB"
EXPIRES="10 years"

# CA
certstrap --depot-path "$OUT" init --common-name root --country "$COUNTRY" --organization "$BRAND" --expires "$EXPIRES" --passphrase ""

# client
certstrap --depot-path "$OUT" request-cert --common-name client --domain client --country "$COUNTRY" --organization "$BRAND" --passphrase ""
certstrap --depot-path "$OUT" sign client --CA root --expires "$EXPIRES"

# server
certstrap --depot-path "$OUT" request-cert --common-name server --domain server --country "$COUNTRY" --organization "$BRAND" --passphrase ""
certstrap --depot-path "$OUT" sign server --CA root --expires "$EXPIRES"
