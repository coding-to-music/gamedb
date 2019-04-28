#!/usr/bin/env bash

# go get github.com/zackslash/goviz

goviz -i github.com/gamedb/gamedb -p | dot -Tpng -o ./imports.png
