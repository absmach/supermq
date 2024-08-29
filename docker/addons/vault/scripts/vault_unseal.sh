#!/usr/bin/bash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

scriptdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
repo_root="$(realpath "$scriptdir/../../../../")"
env_file="$repo_root/docker/.env"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --env-file)
            if [[ "$2" = /* ]]; then
                env_file="$2"
            else
                env_file="$(realpath -m "$repo_root/$2")"
            fi
            shift
            ;;
        *)
            echo "Unknown parameter passed: $1"
            exit 1
            ;;
    esac
    shift
done

readDotEnv() {
    set -o allexport
    source "$env_file"
    set +o allexport
}

source "$scriptdir/vault_cmd.sh"

readDotEnv

vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_1}
vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_2}
vault operator unseal -address=${MG_VAULT_ADDR} ${MG_VAULT_UNSEAL_KEY_3}
