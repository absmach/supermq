#!/bin/bash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

# OpenBao startup and initialization script for SuperMQ
# This script starts OpenBao in dev mode and then sets up the PKI engine, 
# AppRole authentication, and all required roles automatically.

set -e

# Configuration variables with defaults from environment
OPENBAO_ADDR=${BAO_ADDR:-http://localhost:8200}
ROOT_TOKEN=${BAO_DEV_ROOT_TOKEN_ID:-openbao-root-token}
PKI_PATH=${SMQ_CERTS_OPENBAO_PKI_PATH:-pki}
ROLE_NAME=${SMQ_CERTS_OPENBAO_ROLE:-supermq}
APP_ROLE_ID=${SMQ_CERTS_OPENBAO_APP_ROLE:-a70f7f8e-4296-6df9-2ce6-c37517107913}
APP_SECRET=${SMQ_CERTS_OPENBAO_APP_SECRET:-d77bd6b3-45f4-22ce-7831-5967dbcfddce}

echo "Starting OpenBao with automatic initialization for SuperMQ..."

# Start OpenBao in background
echo "Starting OpenBao server..."
bao server -dev -dev-root-token-id="$ROOT_TOKEN" -dev-listen-address="0.0.0.0:8200" &
BAO_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "Shutting down OpenBao..."
    kill $BAO_PID 2>/dev/null || true
    wait $BAO_PID 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Wait for OpenBao to be ready
echo "Waiting for OpenBao to be ready..."
for i in {1..30}; do
    if curl -s -f $OPENBAO_ADDR/v1/sys/health >/dev/null 2>&1; then
        echo "OpenBao is ready!"
        break
    fi
    sleep 2
done

# Setup OpenBao
export BAO_ADDR="$OPENBAO_ADDR"
export BAO_TOKEN="$ROOT_TOKEN"

# Check if already configured
if ! (bao secrets list | grep -q "^${PKI_PATH}/" && bao auth list | grep -q "^approle/"); then
    echo "Configuring OpenBao..."
    
    # Enable auth and secrets
    bao auth enable approle || true
    bao secrets enable -path="$PKI_PATH" pki || true
    bao secrets tune -max-lease-ttl=87600h "$PKI_PATH"
    
    # Generate CA
    bao write -field=certificate "$PKI_PATH/root/generate/internal" common_name="SuperMQ Root CA" ttl=87600h > /tmp/ca.crt || true
    bao write "$PKI_PATH/config/urls" issuing_certificates="$OPENBAO_ADDR/v1/$PKI_PATH/ca" crl_distribution_points="$OPENBAO_ADDR/v1/$PKI_PATH/crl"
    
    # Create role
    bao write "$PKI_PATH/roles/$ROLE_NAME" allow_any_name=true enforce_hostnames=false allow_ip_sans=true allow_localhost=true max_ttl=720h ttl=720h
    
    # Create AppRole
    bao write auth/approle/role/supermq token_policies=pki-policy token_ttl=1h token_max_ttl=4h bind_secret_id=true
    bao write auth/approle/role/supermq/role-id role_id="$APP_ROLE_ID"
    bao write auth/approle/role/supermq/custom-secret-id secret_id="$APP_SECRET"
    
    echo "OpenBao configuration complete!"
    echo "AppRole ID: $APP_ROLE_ID"
    echo "AppRole Secret: $APP_SECRET"
else
    echo "OpenBao already configured, skipping initialization"
fi

echo "OpenBao is ready for SuperMQ!"
wait $BAO_PID
