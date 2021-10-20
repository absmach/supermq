This is Vault service deployment to be used with Mainflux.

When the Vault service is started, some initialization steps need to be done to set things up.

## Configuration

| Variable                  | Description                                                             | Default        |
| ------------------------- | ----------------------------------------------------------------------- | -------------- |
| MF_VAULT_HOST             | Vault service address                                                   | vault          |
| MF_VAULT_PORT             | Vault service port                                                      | 8200           |
| MF_VAULT_UNSEAL_KEY_1     | Vault unseal key                                                        | ""             |
| MF_VAULT_UNSEAL_KEY_2     | Vault unseal key                                                        | ""             |
| MF_VAULT_UNSEAL_KEY_3     | Vault unseal key                                                        | ""             |
| MF_VAULT_TOKEN            | Vault cli access token                                                  | ""             |
| MF_VAULT_PKI_PATH         | Vault secrets engine path for CA                                        | pki            |
| MF_VAULT_PKI_INT_PATH     | Vault secrets engine path for intermediate CA                           | pki_int        |
| MF_VAULT_CA_ROLE_NAME     | Vault secrets engine role                                               | mainflux       |
| MF_VAULT_CA_NAME          | Certificates name used by `vault-set-pki.sh`                            | mainflux       |
| MF_VAULT_CA_CN            | Common name used for CA creation by `vault-set-pki.sh`                  | mainflux.com   |
| MF_VAULT_CA_OU            | Org unit used for CA creation by `vault-set-pki.sh`                     | Mainflux Cloud |
| MF_VAULT_CA_O             | Organization used for CA creation by `vault-set-pki.sh`                 | Mainflux Labs  |
| MF_VAULT_CA_C             | Country used for CA creation by `vault-set-pki.sh`                      | Serbia         |
| MF_VAULT_CA_L             | Location used for CA creation by `vault-set-pki.sh`                     | Belgrade       |


## Setup

The following scripts are provided, which work on the running Vault service in Docker.

1. `init.sh`

This script is execetud in `vault-operator` container. Script waits for vault to be ready and then initializes and unseals `vault`.  
`vault-operator` container uses custom built image `mainflux/vault:latest` see [Dockerfile](./Dockerfile).


Script calls `vault operator init` to perform the initial vault initialization and generates
a `data/secrets` file which contains the Vault unseal keys and root tokens.

This procedure is not production safe, this is only for development.


2. `vault-set-pki.sh`

This script is used to generate the root certificate, intermediate certificate and HTTPS server certificate.
After it runs, it copies the necessary certificates and keys to the `docker/ssl/certs` folder.

The CA parameters as well as vault root token script reads from the environment variables starting with `MF_VAULT_CA` in `.env` file.
So you need to populate `.env` prior to executing this script.

## Vault CLI 

It can also be useful to run the Vault CLI for inspection and administration work.

This can be done directly using the Vault image in Docker: `docker run -it mainflux/vault:latest vault`

```
Usage: vault <command> [args]

Common commands:
    read        Read data and retrieves secrets
    write       Write data, configuration, and secrets
    delete      Delete secrets and configuration
    list        List data or secrets
    login       Authenticate locally
    agent       Start a Vault agent
    server      Start a Vault server
    status      Print seal and HA status
    unwrap      Unwrap a wrapped secret

Other commands:
    audit          Interact with audit devices
    auth           Interact with auth methods
    debug          Runs the debug command
    kv             Interact with Vault's Key-Value storage
    lease          Interact with leases
    monitor        Stream log messages from a Vault server
    namespace      Interact with namespaces
    operator       Perform operator-specific tasks
    path-help      Retrieve API help for paths
    plugin         Interact with Vault plugins and catalog
    policy         Interact with policies
    print          Prints runtime configurations
    secrets        Interact with secrets engines
    ssh            Initiate an SSH session
    token          Interact with tokens
```

### Vault Web UI

The Vault Web UI is accessible by default on `http://localhost:8200/ui`.
