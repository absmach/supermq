#!/usr/bin/bash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

scriptdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd $scriptdir

# Default .env file path
env_file="../../../docker/.env"

# Default certificate copy path
certs_copy_path="../../../docker/ssl/certs/"

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --env-file) env_file="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

readDotEnv() {
    set -o allexport
    source "$env_file"
    set +o allexport
}

readDotEnv

server_name="localhost"

# Check if MG_NGINX_SERVER_NAME is set or not empty
if [ -n "${MG_NGINX_SERVER_NAME:-}" ]; then
    server_name="$MG_NGINX_SERVER_NAME"
fi

echo "Copying certificate files"

if [ -e "data/${server_name}.crt" ]; then
    cp -v data/${server_name}.crt      ${certs_copy_path}magistrala-server.crt
else
    echo "${server_name}.crt file not available"
fi

if [ -e "data/${server_name}.key" ]; then
    cp -v data/${server_name}.key      ${certs_copy_path}magistrala-server.key
else
    echo "${server_name}.key file not available"
fi

if [ -e "data/${MG_VAULT_PKI_INT_FILE_NAME}.key" ]; then
    cp -v data/${MG_VAULT_PKI_INT_FILE_NAME}.key    ${certs_copy_path}ca.key
else
    echo "data/${MG_VAULT_PKI_INT_FILE_NAME}.key file not available"
fi

if [ -e "data/${MG_VAULT_PKI_INT_FILE_NAME}_bundle.crt" ]; then
    cp -v data/${MG_VAULT_PKI_INT_FILE_NAME}_bundle.crt     ${certs_copy_path}ca.crt
else
    echo "data/${MG_VAULT_PKI_INT_FILE_NAME}_bundle.crt file not available"
fi

exit 0
