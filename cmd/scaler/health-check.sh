#!/usr/bin/env bash

response=$(curl --silent http://localhost:8081/health-check || exit 1)

if [[ "${response}" == "OK" ]]; then
  exit 0
else
  exit 1
fi
