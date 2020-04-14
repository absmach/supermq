# PROVISION service

PROVISION service provides an HTTP API to interact with Mainflux.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                            | Description                                       | Default                          |
| ----------------------------------- | ------------------------------------------------- | -------------------------------- |
| MF_PROVISION_USER                   | User (email) for accessing Mainflux               | user@example.com                 |
| MF_PROVISION_PASS                   | Mainflux password                                 | user123                          |
| MF_PROVISION_API_KEY                | Mainflux authentication token                     | ""                               |
| MF_PROVISION_CONFIG_FILE            | Provision config file                             | "config.toml"                    |
| MF_PROVISION_HTTP_PORT              | Provision service listening port                  | 8091                             |
| MF_PROVISION_ENV_CLIENTS_TLS        | Mainflux SDK TLS verification                     | false                            |
| MF_PROVISION_CA_CERTS               | Mainflux gRPC secure certs                        | ""                               |
| MF_PROVISION_SERVER_CERT            | Mainflux gRPC secure server cert                  | ""                               |
| MF_PROVISION_SERVER_KEY             | Mainflux gRPC secure server key                   | ""                               |
| MF_PROVISION_SERVER_KEY             | Mainflux gRPC secure server key                   | ""                               |
| MF_PROVISION_MQTT_URL               | Mainflux MQTT adapter URL                         | "http://localhost:1883"          |
| MF_PROVISION_USERS_LOCATION         | Users service URL                                 | "http://locahost"                |
| MF_PROVISION_THINGS_LOCATION        | Things service URL                                | "http://localhost"               |
| MF_PROVISION_LOG_LEVEL              | Service log level                                 | "http://localhost"               |
| MF_PROVISION_HTTP_PORT              | Service listening port                            | "8091"                           |
| MF_PROVISION_USER                   | Mainflux user username                            | "test@example.com"               |
| MF_PROVISION_PASS                   | Mainflux user password                            | "password"                       |
| MF_PROVISION_BS_SVC_URL             | Mainflux Bootstrap service URL                    | http://localhost/things/configs" |
| MF_PROVISION_BS_SVC_WHITELIST_URL   | Mainflux Bootstrap service whitelist URL          | "http://localhost/things/state"  |
| MF_PROVISION_CERTS_SVC_URL          | Certificats service URL                           | "http://localhost/certs"         |
| MF_PROVISION_X509_PROVISIONING      | Should X509 client cert be provisioned            | "false"                          |
| MF_PROVISION_BS_CONFIG_PROVISIONING | Should thing config be saved in Bootstrap service | "true"                           |
| MF_PROVISION_BS_AUTO_WHITEIST       | Should thing be auto whitelisted                  | "true"                           |
| MF_PROVISION_BS_CONTENT             | Bootstrap service content                         | "{}"                             |

By default, call to mapping endpoint will create one thing and two channels ( 'control' and 'data' ) and connect it. If there is a requirement for different conf
we can use [config](docker/configs/config.toml) file.
In config.toml we can enlist array of things and channels that we want to create and make connections between them.
Metadata can be whatever suits your needs except that at least one thing needs to have 'externalID' (which is populated with value from [request](#example)).
For channels metadata 'type' is reserved for 'control' and 'data'.

Example below
```
[[things]]
  name = "thing"

  [things.metadata]
    externalID = "xxxxxx"


[[channels]]
  name = "control-channel"

  [channels.metadata]
    type = "control"

[[channels]]
  name = "data-channel"

  [channels.metadata]
    type = "data"

[[channels]]
  name = "export-channel"

  [channels.metadata]
    type = "data"
```


## Example 
```
curl -X POST \
  http://localhost:8091/mapping\
  -H 'Content-Type: application/json' \
  -d '{ "externalid" : "02:42:fE:65:CB:3d", "externalkey: "key12345678" }'
  ```
