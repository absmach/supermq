# Clickhouse reader

Clickhouse reader provides message repository implementation for ClickHouse.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                            | Description                                 | Default        |
|-------------------------------------|---------------------------------------------|----------------|
| MF_CLICKHOUSE_READER_LOG_LEVEL        | Service log level                           | debug          |
| MF_CLICKHOUSE_READER_PORT             | Service HTTP port                           | 8180           |
| MF_CLICKHOUSE_READER_CLIENT_TLS       | TLS mode flag                               | false          |
| MF_CLICKHOUSE_READER_CA_CERTS         | Path to trusted CAs in PEM format           |                |
| MF_CLICKHOUSE_READER_DB_HOST          | ClickHouse DB host                            | clickhouse       |
| MF_CLICKHOUSE_READER_DB_PORT          | ClickHouse DB port                            | 5432           |
| MF_CLICKHOUSE_READER_DB_USER          | ClickHouse user                               | mainflux       |
| MF_CLICKHOUSE_READER_DB_PASS          | ClickHouse password                           | mainflux       |
| MF_CLICKHOUSE_READER_DB               | ClickHouse database name                      | messages       |
| MF_CLICKHOUSE_READER_DB_SSL_MODE      | ClickHouse SSL mode                           | disabled       |
| MF_CLICKHOUSE_READER_DB_SSL_CERT      | ClickHouse SSL certificate path               | ""             |
| MF_CLICKHOUSE_READER_DB_SSL_KEY       | ClickHouse SSL key                            | ""             |
| MF_CLICKHOUSE_READER_DB_SSL_ROOT_CERT | ClickHouse SSL root certificate path          | ""             |
| MF_JAEGER_URL                       | Jaeger server URL                           | localhost:6831 |
| MF_THINGS_AUTH_GRPC_URL             | Things service Auth gRPC URL                | localhost:8181 |
| MF_THINGS_AUTH_GRPC_TIMEOUT         | Things service Auth gRPC timeout in seconds | 1              |

## Deployment

```yaml
  version: "3.7"
  clickhouse-writer:
    image: mainflux/clickhouse-writer:[version]
    container_name: [instance name]
    depends_on:
      - clickhouse
      - nats
    restart: on-failure
    environment:
      MF_NATS_URL: [NATS instance URL]
      MF_CLICKHOUSE_READER_LOG_LEVEL: [Service log level]
      MF_CLICKHOUSE_READER_PORT: [Service HTTP port]
      MF_CLICKHOUSE_READER_DB_HOST: [ClickHouse host]
      MF_CLICKHOUSE_READER_DB_PORT: [ClickHouse port]
      MF_CLICKHOUSE_READER_DB_USER: [ClickHouse user]
      MF_CLICKHOUSE_READER_DB_PASS: [ClickHouse password]
      MF_CLICKHOUSE_READER_DB: [ClickHouse database name]
      MF_CLICKHOUSE_READER_DB_SSL_MODE: [ClickHouse SSL mode]
      MF_CLICKHOUSE_READER_DB_SSL_CERT: [ClickHouse SSL cert]
      MF_CLICKHOUSE_READER_DB_SSL_KEY: [ClickHouse SSL key]
      MF_CLICKHOUSE_READER_DB_SSL_ROOT_CERT: [ClickHouse SSL Root cert]
      MF_JAEGER_URL: [Jaeger server URL]
      MF_THINGS_AUTH_GRPC_URL: [Things service Auth gRPC URL]
      MF_THINGS_AUTH_GRPC_TIMEOUT: [Things service Auth gRPC request timeout in seconds]
    ports:
      - 8180:8180
    networks:
      - docker_mainflux-base-net
```

To start the service, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/mainflux/mainflux

cd mainflux

# compile the clickhouse writer
make clickhouse-writer

# copy binary to bin
make install

# Set the environment variables and run the service
MF_CLICKHOUSE_READER_LOG_LEVEL=[Service log level] \
MF_CLICKHOUSE_READER_PORT=[Service HTTP port] \
MF_CLICKHOUSE_READER_CLIENT_TLS =[TLS mode flag] \
MF_CLICKHOUSE_READER_CA_CERTS=[Path to trusted CAs in PEM format] \
MF_CLICKHOUSE_READER_DB_HOST=[ClickHouse host] \
MF_CLICKHOUSE_READER_DB_PORT=[ClickHouse port] \
MF_CLICKHOUSE_READER_DB_USER=[ClickHouse user] \
MF_CLICKHOUSE_READER_DB_PASS=[ClickHouse password] \
MF_CLICKHOUSE_READER_DB=[ClickHouse database name] \
MF_CLICKHOUSE_READER_DB_SSL_MODE=[ClickHouse SSL mode] \
MF_CLICKHOUSE_READER_DB_SSL_CERT=[ClickHouse SSL cert] \
MF_CLICKHOUSE_READER_DB_SSL_KEY=[ClickHouse SSL key] \
MF_CLICKHOUSE_READER_DB_SSL_ROOT_CERT=[ClickHouse SSL Root cert] \
MF_JAEGER_URL=[Jaeger server URL] \
MF_THINGS_AUTH_GRPC_URL=[Things service Auth GRPC URL] \
MF_THINGS_AUTH_GRPC_TIMEOUT=[Things service Auth gRPC request timeout in seconds] \
$GOBIN/mainflux-clickhouse-reader
```

## Usage

Starting service will start consuming normalized messages in SenML format.
