#!/usr/bin/env bash

go get github.com/fzipp/gocyclo/cmd/gocyclo

gocyclo ../
#gocyclo ../main.go
#gocyclo -top 10 ../src/
#gocyclo -over 25 ../docker
#gocyclo -avg ../
#gocyclo -top 20 -ignore "_test|Godeps|vendor/" ../.
#gocyclo -over 3 -avg ../gocyclo/
