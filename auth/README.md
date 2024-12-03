# Auth - Authentication and Authorization service

Auth service provides authentication features as an API for managing authentication keys as well as administering groups of entities - `clients` and `users`.

## Authentication

User service is using Auth service gRPC API to obtain login token or password reset token. Authentication key consists of the following fields:

- ID - key ID
- Type - one of the three types described below
- IssuerID - an ID of the SuperMQ User who issued the key
- Subject - user ID for which the key is issued
- IssuedAt - the timestamp when the key is issued
- ExpiresAt - the timestamp after which the key is invalid

There are four types of authentication keys:

- Access key - keys issued to the user upon login request
- Refresh key - keys used to generate new access keys
- Recovery key - password recovery key
- API key - keys issued upon the user request
- Invitation key - keys used to invite new users

Authentication keys are represented and distributed by the corresponding [JWT](jwt.io).

User keys are issued when user logs in. Each user request (other than `registration` and `login`) contains user key that is used to authenticate the user.

API keys are similar to the User keys. The main difference is that API keys have configurable expiration time. If no time is set, the key will never expire. For that reason, API keys are _the only key type that can be revoked_. This also means that, despite being used as a JWT, it requires a query to the database to validate the API key. The user with API key can perform all the same actions as the user with login key (can act on behalf of the user for Client, Channel, or user profile management), _except issuing new API keys_.

Recovery key is the password recovery key. It's short-lived token used for password recovery process.

For in-depth explanation of the aforementioned scenarios, as well as thorough understanding of SuperMQ, please check out the [official documentation][doc].

The following actions are supported:

- create (all key types)
- verify (all key types)
- obtain (API keys only)
- revoke (API keys only)

## Domains

Domains are used to group users and clients. Each domain has a unique alias that is used to identify the domain. Domains are used to group users and their entities.

Domain consists of the following fields:

- ID - UUID uniquely representing domain
- Name - name of the domain
- Tags - array of tags
- Metadata - Arbitrary, object-encoded domain's data
- Alias - unique alias of the domain
- CreatedAt - timestamp at which the domain is created
- UpdatedAt - timestamp at which the domain is updated
- UpdatedBy - user that updated the domain
- CreatedBy - user that created the domain
- Status - domain status

## Configuration

The service is configured using the environment variables presented in the following table. Note that any unset variables will be replaced with their default values.

| Variable                        | Description                                                             | Default                        |
| ------------------------------- | ----------------------------------------------------------------------- | ------------------------------ |
| SMQ_AUTH_LOG_LEVEL              | Log level for the Auth service (debug, info, warn, error)               | info                           |
| SMQ_AUTH_DB_HOST                | Database host address                                                   | localhost                      |
| SMQ_AUTH_DB_PORT                | Database host port                                                      | 5432                           |
| SMQ_AUTH_DB_USER                | Database user                                                           | supermq                        |
| SMQ_AUTH_DB_PASSWORD            | Database password                                                       | supermq                        |
| SMQ_AUTH_DB_NAME                | Name of the database used by the service                                | auth                           |
| SMQ_AUTH_DB_SSL_MODE            | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable                        |
| SMQ_AUTH_DB_SSL_CERT            | Path to the PEM encoded certificate file                                | ""                             |
| SMQ_AUTH_DB_SSL_KEY             | Path to the PEM encoded key file                                        | ""                             |
| SMQ_AUTH_DB_SSL_ROOT_CERT       | Path to the PEM encoded root certificate file                           | ""                             |
| SMQ_AUTH_HTTP_HOST              | Auth service HTTP host                                                  | ""                             |
| SMQ_AUTH_HTTP_PORT              | Auth service HTTP port                                                  | 8189                           |
| SMQ_AUTH_HTTP_SERVER_CERT       | Path to the PEM encoded HTTP server certificate file                    | ""                             |
| SMQ_AUTH_HTTP_SERVER_KEY        | Path to the PEM encoded HTTP server key file                            | ""                             |
| SMQ_AUTH_GRPC_HOST              | Auth service gRPC host                                                  | ""                             |
| SMQ_AUTH_GRPC_PORT              | Auth service gRPC port                                                  | 8181                           |
| SMQ_AUTH_GRPC_SERVER_CERT       | Path to the PEM encoded gRPC server certificate file                    | ""                             |
| SMQ_AUTH_GRPC_SERVER_KEY        | Path to the PEM encoded gRPC server key file                            | ""                             |
| SMQ_AUTH_GRPC_SERVER_CA_CERTS   | Path to the PEM encoded gRPC server CA certificate file                 | ""                             |
| SMQ_AUTH_GRPC_CLIENT_CA_CERTS   | Path to the PEM encoded gRPC client CA certificate file                 | ""                             |
| SMQ_AUTH_SECRET_KEY             | String used for signing tokens                                          | secret                         |
| SMQ_AUTH_ACCESS_TOKEN_DURATION  | The access token expiration period                                      | 1h                             |
| SMQ_AUTH_REFRESH_TOKEN_DURATION | The refresh token expiration period                                     | 24h                            |
| SMQ_AUTH_INVITATION_DURATION    | The invitation token expiration period                                  | 168h                           |
| SMQ_SPICEDB_HOST                | SpiceDB host address                                                    | localhost                      |
| SMQ_SPICEDB_PORT                | SpiceDB host port                                                       | 50051                          |
| SMQ_SPICEDB_PRE_SHARED_KEY      | SpiceDB pre-shared key                                                  | 12345678                       |
| SMQ_SPICEDB_SCHEMA_FILE         | Path to SpiceDB schema file                                             | ./docker/spicedb/schema.zed    |
| SMQ_JAEGER_URL                  | Jaeger server URL                                                       | <http://jaeger:4318/v1/traces> |
| SMQ_JAEGER_TRACE_RATIO          | Jaeger sampling ratio                                                   | 1.0                            |
| SMQ_SEND_TELEMETRY              | Send telemetry to supermq call home server                              | true                           |
| SMQ_AUTH_ADAPTER_INSTANCE_ID    | Adapter instance ID                                                     | ""                             |

## Deployment

The service itself is distributed as Docker container. Check the [`auth`](https://github.com/absmach/supermq/blob/main/docker/docker-compose.yml) service section in docker-compose file to see how service is deployed.

Running this service outside of container requires working instance of the postgres database, SpiceDB, and Jaeger server.
To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/absmach/supermq

cd supermq

# compile the service
make auth

# copy binary to bin
make install

# set the environment variables and run the service
SMQ_AUTH_LOG_LEVEL=info \
SMQ_AUTH_DB_HOST=localhost \
SMQ_AUTH_DB_PORT=5432 \
SMQ_AUTH_DB_USER=supermq \
SMQ_AUTH_DB_PASSWORD=supermq \
SMQ_AUTH_DB_NAME=auth \
SMQ_AUTH_DB_SSL_MODE=disable \
SMQ_AUTH_DB_SSL_CERT="" \
SMQ_AUTH_DB_SSL_KEY="" \
SMQ_AUTH_DB_SSL_ROOT_CERT="" \
SMQ_AUTH_HTTP_HOST=localhost \
SMQ_AUTH_HTTP_PORT=8189 \
SMQ_AUTH_HTTP_SERVER_CERT="" \
SMQ_AUTH_HTTP_SERVER_KEY="" \
SMQ_AUTH_GRPC_HOST=localhost \
SMQ_AUTH_GRPC_PORT=8181 \
SMQ_AUTH_GRPC_SERVER_CERT="" \
SMQ_AUTH_GRPC_SERVER_KEY="" \
SMQ_AUTH_GRPC_SERVER_CA_CERTS="" \
SMQ_AUTH_GRPC_CLIENT_CA_CERTS="" \
SMQ_AUTH_SECRET_KEY=secret \
SMQ_AUTH_ACCESS_TOKEN_DURATION=1h \
SMQ_AUTH_REFRESH_TOKEN_DURATION=24h \
SMQ_AUTH_INVITATION_DURATION=168h \
SMQ_SPICEDB_HOST=localhost \
SMQ_SPICEDB_PORT=50051 \
SMQ_SPICEDB_PRE_SHARED_KEY=12345678 \
SMQ_SPICEDB_SCHEMA_FILE=./docker/spicedb/schema.zed \
SMQ_JAEGER_URL=http://localhost:14268/api/traces \
SMQ_JAEGER_TRACE_RATIO=1.0 \
SMQ_SEND_TELEMETRY=true \
SMQ_AUTH_ADAPTER_INSTANCE_ID="" \
$GOBIN/supermq-auth
```

Setting `SMQ_AUTH_HTTP_SERVER_CERT` and `SMQ_AUTH_HTTP_SERVER_KEY` will enable TLS against the service. The service expects a file in PEM format for both the certificate and the key.
Setting `SMQ_AUTH_GRPC_SERVER_CERT` and `SMQ_AUTH_GRPC_SERVER_KEY` will enable TLS against the service. The service expects a file in PEM format for both the certificate and the key. Setting `SMQ_AUTH_GRPC_SERVER_CA_CERTS` will enable TLS against the service trusting only those CAs that are provided. The service expects a file in PEM format of trusted CAs. Setting `SMQ_AUTH_GRPC_CLIENT_CA_CERTS` will enable TLS against the service trusting only those CAs that are provided. The service expects a file in PEM format of trusted CAs.

## Usage

For more information about service capabilities and its usage, please check out the [API documentation](https://docs.api.supermq.abstractmachines.fr/?urls.primaryName=auth.yml).

[doc]: https://docs.supermq.abstractmachines.fr
