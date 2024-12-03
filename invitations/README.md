# Invitation Service

Invitation service is responsible for sending invitations to users to join a domain.

## Configuration

The service is configured using the environment variables presented in the following table. Note that any unset variables will be replaced with their default values.

| Variable                        | Description                                      | Default                 |
| ------------------------------- | ------------------------------------------------ | ----------------------- |
| SMQ_INVITATION_LOG_LEVEL         | Log level for the Invitation service             | debug                   |
| SMQ_USERS_URL                    | Users service URL                                | <http://localhost:9002> |
| SMQ_DOMAINS_URL                  | Domains service URL                              | <http://localhost:8189> |
| SMQ_INVITATIONS_HTTP_HOST        | Invitation service HTTP listening host           | localhost               |
| SMQ_INVITATIONS_HTTP_PORT        | Invitation service HTTP listening port           | 9020                    |
| SMQ_INVITATIONS_HTTP_SERVER_CERT | Invitation service server certificate            | ""                      |
| SMQ_INVITATIONS_HTTP_SERVER_KEY  | Invitation service server key                    | ""                      |
| SMQ_AUTH_GRPC_URL                | Auth service gRPC URL                            | localhost:8181          |
| SMQ_AUTH_GRPC_TIMEOUT            | Auth service gRPC request timeout in seconds     | 1s                      |
| SMQ_AUTH_GRPC_CLIENT_CERT        | Path to client certificate in PEM format         | ""                      |
| SMQ_AUTH_GRPC_CLIENT_KEY         | Path to client key in PEM format                 | ""                      |
| SMQ_AUTH_GRPC_CLIENT_CA_CERTS    | Path to trusted CAs in PEM format                | ""                      |
| SMQ_INVITATIONS_DB_HOST          | Invitation service database host                 | localhost               |
| SMQ_INVITATIONS_DB_USER          | Invitation service database user                 | supermq              |
| SMQ_INVITATIONS_DB_PASS          | Invitation service database password             | supermq              |
| SMQ_INVITATIONS_DB_PORT          | Invitation service database port                 | 5432                    |
| SMQ_INVITATIONS_DB_NAME          | Invitation service database name                 | invitations             |
| SMQ_INVITATIONS_DB_SSL_MODE      | Invitation service database SSL mode             | disable                 |
| SMQ_INVITATIONS_DB_SSL_CERT      | Invitation service database SSL certificate      | ""                      |
| SMQ_INVITATIONS_DB_SSL_KEY       | Invitation service database SSL key              | ""                      |
| SMQ_INVITATIONS_DB_SSL_ROOT_CERT | Invitation service database SSL root certificate | ""                      |
| SMQ_INVITATIONS_INSTANCE_ID      | Invitation service instance ID                   |                         |

## Deployment

The service itself is distributed as Docker container. Check the [`invitation`](https://github.com/absmach/amdm/blob/main/docker/docker-compose.yml) service section in docker-compose file to see how service is deployed.

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/absmach/supermq

cd supermq

# compile the http
make invitation

# copy binary to bin
make install

# set the environment variables and run the service
SMQ_INVITATION_LOG_LEVEL=info \
SMQ_INVITATIONS_ENDPOINT=/invitations \
SMQ_USERS_URL="http://localhost:9002" \
SMQ_DOMAINS_URL="http://localhost:8189" \
SMQ_INVITATIONS_HTTP_HOST=localhost \
SMQ_INVITATIONS_HTTP_PORT=9020 \
SMQ_INVITATIONS_HTTP_SERVER_CERT="" \
SMQ_INVITATIONS_HTTP_SERVER_KEY="" \
SMQ_AUTH_GRPC_URL=localhost:8181 \
SMQ_AUTH_GRPC_TIMEOUT=1s \
SMQ_AUTH_GRPC_CLIENT_CERT="" \
SMQ_AUTH_GRPC_CLIENT_KEY="" \
SMQ_AUTH_GRPC_CLIENT_CA_CERTS="" \
SMQ_INVITATIONS_DB_HOST=localhost \
SMQ_INVITATIONS_DB_USER=supermq \
SMQ_INVITATIONS_DB_PASS=supermq \
SMQ_INVITATIONS_DB_PORT=5432 \
SMQ_INVITATIONS_DB_NAME=invitations \
SMQ_INVITATIONS_DB_SSL_MODE=disable \
SMQ_INVITATIONS_DB_SSL_CERT="" \
SMQ_INVITATIONS_DB_SSL_KEY="" \
SMQ_INVITATIONS_DB_SSL_ROOT_CERT="" \
$GOBIN/supermq-invitation
```

## Usage

For more information about service capabilities and its usage, please check out the [API documentation](https://docs.api.supermq.abstractmachines.fr/?urls.primaryName=invitations.yml).
