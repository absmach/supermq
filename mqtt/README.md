# MQTT adapter

MQTT adapter provides an MQTT API for sending and receiving messages through the 
platform.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable             | Description         | Default               |
|----------------------|---------------------|-----------------------|
| MF_MQTT_ADAPTER_PORT | Service MQTT port   | 1883                  |
| MF_MQTT_WS_PORT      | WebSocket port      | 8880                  |
| MF_NATS_URL          | NATS instance URL   | nats://localhost:4222 |
| MF_THINGS_URL        | Things service URL  | localhost:8181        |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  mqtt:
    image: mainflux/mqtt:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured port]
    environment:
      MF_THINGS_URL: [Things service URL]
      MF_NATS_URL: [NATS instance URL]
      MF_MQTT_ADAPTER_PORT: [Service MQTT port]
      MF_MQTT_WS_PORT: [Service WS port]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux/mqtt

# install dependencies
npm install

# set the environment variables and run the service
MF_THINGS_URL=[Things service URL] MF_NATS_URL=[NATS instance URL] MF_MQTT_ADAPTER_PORT=[Service MQTT port] MF_MQTT_WS_PORT= [Service WS port] node mqtt.js
```

## Usage

To use MQTT adapter you should use `channels/<channel_id>/messages`. Client key should
be passed as user's password.
