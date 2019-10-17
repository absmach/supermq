# OPC-UA Adapter
Adapter between Mainflux IoT system and an OPC-UA Server.

This adapter sits between Mainflux and an OPC-UA server and just forwards the messages from one system to another.

OPC-UA Server is used for connectivity layer and the data is pushed via this adapter service to Mainflux, where it is persisted and routed to other protocols via Mainflux multi-protocol message broker. Mainflux adds user accounts, application management and security in order to obtain the overall end-to-end OPC-UA solution.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                        | Description                           | Default               |
|---------------------------------|---------------------------------------|-----------------------|
| MF_OPC_ADAPTER_HTTP_PORT        | Service HTTP port                     | 8180                  |
| MF_OPC_ADAPTER_LOG_LEVEL        | Log level for the OPC-UA Adapter      | error                 |
| MF_NATS_URL                     | NATS instance URL                     | nats://localhost:4222 |
| MF_OPC_ADAPTER_MESSAGES_URL     | OPC-UA Server mqtt broker URL         | tcp://localhost:1883  |
| MF_OPC_ADAPTER_ROUTEMAP_URL     | Routemap database URL                 | localhost:6379        |
| MF_OPC_ADAPTER_ROUTEMAP_PASS    | Routemap database password            |                       |
| MF_OPC_ADAPTER_ROUTEMAP_DB      | Routemap instance that should be used | 0                     |
| MF_THINGS_ES_URL                | Things service event store URL        | localhost:6379        |
| MF_THINGS_ES_PASS               | Things service event store password   |                       |
| MF_THINGS_ES_DB                 | Things service event store db         | 0                     |
| MF_OPC_ADAPTER_INSTANCE_NAME    | OPC-UA adapter instance name          | opc                   |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  adapter:
    image: mainflux/opc:[version]
    container_name: [instance name]
    environment:
      MF_OPC_ADAPTER_LOG_LEVEL: [OPC-UA Adapter Log Level]
      MF_NATS_URL: [NATS instance URL]
      MF_OPC_ADAPTER_MESSAGES_URL: [OPC-UA Server mqtt broker URL]
      MF_OPC_ADAPTER_ROUTEMAP_URL: [OPC-UA adapter routemap URL]
      MF_OPC_ADAPTER_ROUTEMAP_PASS: [OPC-UA adapter routemap password]
      MF_OPC_ADAPTER_ROUTEMAP_DB: [OPC-UA adapter routemap instance]
      MF_THINGS_ES_URL: [Things service event store URL]
      MF_THINGS_ES_PASS: [Things service event store password]
      MF_THINGS_ES_DB: [Things service event store db]
      MF_OPC_ADAPTER_INSTANCE_NAME: [OPC-UA adapter instance name]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the opc adapter
make opcua

# copy binary to bin
make install

# set the environment variables and run the service
MF_OPC_ADAPTER_LOG_LEVEL=[OPC-UA Adapter Log Level] MF_NATS_URL=[NATS instance URL] MF_OPC_ADAPTER_MESSAGES_URL=[OPC-UA Server mqtt broker URL] MF_OPC_ADAPTER_ROUTEMAP_URL=[OPC-UA adapter routemap URL] MF_OPC_ADAPTER_ROUTEMAP_PASS=[OPC-UA adapter routemap password] MF_OPC_ADAPTER_ROUTEMAP_DB=[OPC-UA adapter routemap instance] MF_THINGS_ES_URL=[Things service event store URL] MF_THINGS_ES_PASS=[Things service event store password] MF_THINGS_ES_DB=[Things service event store db] MF_OPC_ADAPTER_INSTANCE_NAME=[OPC-UA adapter instance name] $GOBIN/mainflux-opc
```

### Using docker-compose

This service can be deployed using docker containers.
Docker compose file is available in `<project_root>/docker/addons/opcua-adapter/docker-compose.yml`. In order to run Mainflux opcua-adapter, execute the following command:

```bash
docker-compose -f docker/addons/opcua-adapter/docker-compose.yml up -d
```

## Usage

For more information about service capabilities and its usage, please check out
the [Mainflux documentation](https://mainflux.readthedocs.io/en/latest/opc/).
