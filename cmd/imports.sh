#!/usr/bin/env bash

# go get github.com/zackslash/goviz

goviz -i github.com/gamedb/website -p | dot -Tpng -o ./imports.png
