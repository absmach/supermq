# LoRa Adapter
Adapter between Mainflux IoT system and [LoRa Server](https://github.com/brocaar/loraserver).

This adapter sits between Mainflux and LoRa server and just forwards the messages form one system to another via MQTT protocol, using the adequate MQTT topics and in the good message format (JSON and SenML), i.e. respecting the APIs of both systems.

LoRa Server is used for connectivity layer and data is pushed via this adapter service to Mainflux, where it is persisted and routed to other protocols via Mainflux multi-protocol message broker. Mainflux adds user accounts, application management and security in order to obtain the overall end-to-end LoRa solution.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                         | Description                           | Default               |
|----------------------------------|---------------------------------------|-----------------------|
| MF_LORA_ADAPTER_LOG_LEVEL        | Log level for the Lora Adapter        | error                 |
| MF_NATS_URL                      | NATS instance URL                     | nats://localhost:4222 |
| MF_LORA_ADAPTER_LORA_MESSAGE_URL | Loraserver mqtt broker URL            | tcp://localhost:1883  |
| MF_LORA_ADAPTER_LORA_SERVER_URL  | Loraserver gRPC API URL               | localhost:8080        |
| MF_LORA_ADAPTER_ROUTEMAP_URL     | Routemap database URL                 | localhost:6379        |
| MF_LORA_ADAPTER_ROUTEMAP_PASS    | Routemap database password            |                       |
| MF_LORA_ADAPTER_ROUTEMAP_DB      | Routemap instance that should be used | 0                     |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  adapter:
    image: mainflux/lora:[version]
    container_name: [instance name]
    environment:
      MF_LORA_ADAPTER_LOG_LEVEL: [Lora Adapter Log Level]
      MF_NATS_URL: [NATS instance URL]
      MF_LORA_ADAPTER_LORA_MESSAGE_URL: [Loraserver mqtt broker URL]
      MF_LORA_ADAPTER_LORA_SERVER_URL: [Loraserver gRPC API URL]
      MF_LORA_ADAPTER_ROUTEMAP_URL: [Lora adapter routemap URL]
      MF_LORA_ADAPTER_ROUTEMAP_PASS: [Lora adapter routemap password]
      MF_LORA_ADAPTER_ROUTEMAP_DB: [Lora adapter routemap instance]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the lora adapter
make lora

# copy binary to bin
make install

# set the environment variables and run the service
MF_LORA_ADAPTER_LOG_LEVEL=[Lora Adapter Log Level] MF_NATS_URL=[NATS instance URL] MF_LORA_ADAPTER_LORA_MESSAGE_URL=[Loraserver mqtt broker URL] MF_LORA_ADAPTER_LORA_SERVER_URL=[Loraserver gRPC API URL] MF_LORA_ADAPTER_ROUTEMAP_URL=[Lora adapter routemap URL] MF_LORA_ADAPTER_ROUTEMAP_PASS=[Lora adapter routemap password] MF_LORA_ADAPTER_ROUTEMAP_DB=[Lora adapter routemap instance] $GOBIN/mainflux-lora
```

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).
