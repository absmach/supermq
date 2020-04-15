# Clickhouse writer

Clickhouse writer provides message repository implementation for ClickHouse.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                             | Description                                 | Default                |
|--------------------------------------|---------------------------------------------|------------------------|
| MF_NATS_URL                          | NATS instance URL                           | nats://localhost:4222  |
| MF_CLICKHOUSE_WRITER_LOG_LEVEL         | Service log level                           | error                  |
| MF_CLICKHOUSE_WRITER_PORT              | Service HTTP port                           | 9000                   |
| MF_CLICKHOUSE_WRITER_DB_HOST           | ClickHouse DB host                            | clickhouse               |
| MF_CLICKHOUSE_WRITER_DB_PORT           | ClickHouse DB port                            | 5432                   |
| MF_CLICKHOUSE_WRITER_DB_USER           | ClickHouse user                               | mainflux               |
| MF_CLICKHOUSE_WRITER_DB_PASS           | ClickHouse password                           | mainflux               |
| MF_CLICKHOUSE_WRITER_DB_NAME           | ClickHouse database name                      | messages               |
| MF_CLICKHOUSE_WRITER_SUBJECTS_CONFIG   | Configuration file path with subjects list  | /config/subjects.toml  |

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
      MF_CLICKHOUSE_WRITER_LOG_LEVEL: [Service log level]
      MF_CLICKHOUSE_WRITER_PORT: [Service HTTP port]
      MF_CLICKHOUSE_WRITER_DB_HOST: [ClickHouse host]
      MF_CLICKHOUSE_WRITER_DB_PORT: [ClickHouse port]
      MF_CLICKHOUSE_WRITER_DB_USER: [ClickHouse user]
      MF_CLICKHOUSE_WRITER_DB_PASS: [ClickHouse password]
      MF_CLICKHOUSE_WRITER_DB_NAME: [ClickHouse database name]
      MF_CLICKHOUSE_WRITER_SUBJECTS_CONFIG: [Configuration file path with subjects list]
    ports:
      - 9000:9000
    networks:
      - docker_mainflux-base-net
    volume:
      - ./subjects.yaml:/config/subjects.yaml
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
MF_NATS_URL=[NATS instance URL] MF_CLICKHOUSE_WRITER_LOG_LEVEL=[Service log level] MF_CLICKHOUSE_WRITER_PORT=[Service HTTP port] MF_CLICKHOUSE_WRITER_DB_HOST=[ClickHouse host] MF_CLICKHOUSE_WRITER_DB_PORT=[ClickHouse port] MF_CLICKHOUSE_WRITER_DB_USER=[ClickHouse user] MF_CLICKHOUSE_WRITER_DB_PASS=[ClickHouse password] MF_CLICKHOUSE_WRITER_DB_NAME=[ClickHouse database name] MF_CLICKHOUSE_WRITER_SUBJECTS_CONFIG=[Configuration file path with subjects list] $GOBIN/mainflux-clickhouse-writer
```

## Usage

Starting service will start consuming normalized messages in SenML format.
