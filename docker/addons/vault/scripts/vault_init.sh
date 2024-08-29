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

source "$repo_root/docker/addons/vault/scripts/vault_cmd.sh"

readDotEnv

mkdir -p "$repo_root/docker/addons/vault/scripts/data"

vault operator init -address="$MG_VAULT_ADDR" 2>&1 | tee >(sed -r 's/\x1b\[[0-9;]*m//g' > "$repo_root/docker/addons/vault/scripts/data/secrets")
