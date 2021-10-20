#!/usr/bin/dumb-init /bin/sh

VAULT_CONFIG_DIR=/vault/config

docker-entrypoint.sh server &
VAULT_PID=$!

sleep 2

wait $VAULT_PID