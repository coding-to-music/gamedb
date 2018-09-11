#!/bin/sh

curl --basic --user ${STEAM_ADMIN_USER}:${STEAM_ADMIN_PASS} --silent "$STEAM_DOMAIN/admin/ranks"
