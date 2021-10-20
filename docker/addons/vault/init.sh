#!/usr/local/bin/bash
# vault status response
# {
#   "type": "shamir",
#   "initialized": true,
#   "sealed": true,
#   "t": 3,
#   "n": 5,
#   "progress": 0,
#   "nonce": "",
#   "version": "1.6.2",
#   "migration": false,
#   "recovery_seal": false,
#   "storage_type": "file",
#   "ha_enabled": false,
#   "active_time": "0001-01-01T00:00:00Z"
# }

# {
#   "unseal_keys_b64": [
#     "63zeIy//W3jLFXFl5rNR+hzRVRpY2lnUVKpilivebKp0",
#     "x9lTCH6DzV/gKkGD+A3EnXDM2I+OmO5INm8RBExlLDFa",
#     "tqo/J6XIgxVtfpsFA29WP07wJoORL/kXfXOMTwo4Wmry",
#     "HxmbF2tTse8eES7ibtPxdQ9JUdMRwz3NXMdeycr7HbS4",
#     "QuYCiO7gtTm3xSgeDqZbsdHffappGAfD5g7lOx1x3LMz"
#   ],
#   "unseal_keys_hex": [
#     "eb7cde232fff5b78cb157165e6b351fa1cd1551a58da59d454aa62962bde6caa74",
#     "c7d953087e83cd5fe02a4183f80dc49d70ccd88f8e98ee48366f11044c652c315a",
#     "b6aa3f27a5c883156d7e9b05036f563f4ef02683912ff9177d738c4f0a385a6af2",
#     "1f199b176b53b1ef1e112ee26ed3f1750f4951d311c33dcd5cc75ec9cafb1db4b8",
#     "42e60288eee0b539b7c5281e0ea65bb1d1df7daa691807c3e60ee53b1d71dcb333"
#   ],
#   "unseal_shares": 5,
#   "unseal_threshold": 3,
#   "recovery_keys_b64": [],
#   "recovery_keys_hex": [],
#   "recovery_keys_shares": 5,
#   "recovery_keys_threshold": 3,
#   "root_token": "s.s8QwAUZcCXDnHdVt4ThJ58z8"
# }
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
        unseal_key1=`echo $secrets | jq -r '.unseal_keys_hex[0]'`
        unseal_key2=`echo $secrets | jq -r '.unseal_keys_hex[1]'`
        unseal_key3=`echo $secrets | jq -r '.unseal_keys_hex[2]'`

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
