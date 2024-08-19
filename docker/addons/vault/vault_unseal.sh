#!/usr/bin/bash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

scriptdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd $scriptdir

# Default .env file path
env_file="../../../docker/.env"

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

source vault_cmd.sh

readDotEnv

vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_1}
vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_2}
vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_3}
