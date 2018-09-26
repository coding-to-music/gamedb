#!/usr/bin/env bash

# Ban everything
docker exec -it varnish varnishadm ban "req.url ~ /"
docker exec -it varnish varnishadm ban.list
