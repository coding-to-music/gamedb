#!/usr/bin/env bash

for BIN in chatbot consumers scaler steam "test" webserver; do

  cd ./cmd/$BIN || exit
  go get -u
  cd ../../

done
