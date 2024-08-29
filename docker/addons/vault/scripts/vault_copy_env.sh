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

write_env() {
    data_dir="$scriptdir/data"

    if [ -e "$data_dir/secrets" ]; then
        sed -i "s,MG_VAULT_UNSEAL_KEY_1=.*,MG_VAULT_UNSEAL_KEY_1=$(awk -F ': ' '$1 == \"Unseal Key 1\" {print $2}' "$data_dir/secrets")," "$env_file"
        sed -i "s,MG_VAULT_UNSEAL_KEY_2=.*,MG_VAULT_UNSEAL_KEY_2=$(awk -F ': ' '$1 == \"Unseal Key 2\" {print $2}' "$data_dir/secrets")," "$env_file"
        sed -i "s,MG_VAULT_UNSEAL_KEY_3=.*,MG_VAULT_UNSEAL_KEY_3=$(awk -F ': ' '$1 == \"Unseal Key 3\" {print $2}' "$data_dir/secrets")," "$env_file"
        sed -i "s,MG_VAULT_TOKEN=.*,MG_VAULT_TOKEN=$(awk -F ': ' '$1 == \"Initial Root Token\" {print $2}' "$data_dir/secrets")," "$env_file"
        echo "Vault environment variables are set successfully in $env_file"
    else
        echo "Error: Source file '$data_dir/secrets' not found."
    fi
}

write_env