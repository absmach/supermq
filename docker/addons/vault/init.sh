#!/usr/local/bin/bash

/wait &&
root_token="not set"
vault_status=`vault status --format=json`

if [ "$vault_status" = "" ]; then
    echo "Failed to check vault status"
    exit 1
fi

vault_initialized=`echo $vault_status | jq '.initialized'`
vault_sealed=`echo $vault_status | jq ' .sealed'`


if [ $vault_initialized = "true" ]; 
then
    echo "vault is initialized"
    if [ $vault_sealed = "true" ];
    then
        echo "vault is sealed"
        secrets=`cat data/secrets`
        unseal_key1=`echo $secrets | jq -r '.unseal_keys_b64[0]'`
        unseal_key2=`echo $secrets | jq -r '.unseal_keys_b64[1]'`
        unseal_key3=`echo $secrets | jq -r '.unseal_keys_b64[2]'`

        vault operator unseal $unseal_key1
        vault operator unseal $unseal_key2
        vault operator unseal $unseal_key3
        root_token=`echo $secrets | jq '.root_token'`
    else
        echo "vault is unsealed, ready to be used"
        exit 0
    fi
else
    echo "initialize vault"
    init_response=`vault operator init -format=json`
    echo $init_response >  data/secrets
    unseal_key1=`echo $init_response | jq -r '.unseal_keys_b64[0]'`
    unseal_key2=`echo $init_response | jq -r '.unseal_keys_b64[1]'`
    unseal_key3=`echo $init_response | jq -r '.unseal_keys_b64[2]'`

    vault operator unseal "$unseal_key1"
    vault operator unseal "$unseal_key2"
    vault operator unseal "$unseal_key3"
    root_token=`echo $init_response | jq '.root_token'`
    sleep 3
fi


vault_initialized=`vault status --format=json | jq '. | "\(.initialized)"'`

echo "root token: $root_token"
