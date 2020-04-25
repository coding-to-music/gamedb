#!/usr/bin/env bash

goda graph ../...:root | dot -Tsvg -o dependencies.svg
