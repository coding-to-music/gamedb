#!/usr/bin/env bash

#cd ~/
#go get -u github.com/loov/goda
#go get -u github.com/goccy/go-graphviz
#brew install graphviz
#cd -

cd ../

goda graph ./...:root | dot -Tsvg -o goda.svg
