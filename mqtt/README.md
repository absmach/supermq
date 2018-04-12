# MQTT adapter

MQTT adapter provides an MQTT API for sending and receiving messages through the platform.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable             | Description         | Default               |
|----------------------|---------------------|-----------------------|
| MF_MANAGER_URL       | Manager service URL | http://localhost:8180 |
| MF_NATS_URL          | NATS instance URL   | nats://localhost:4222 |
| MF_MQTT_ADAPTER_PORT | Service MQTT port   | 1883                  |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  adapter:
    image: mainflux/mqtt:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured port]
    environment:
      MF_MANAGER_URL: [Manager service URL]
      MF_NATS_URL: [NATS instance URL]
      MF_MQTT_ADAPTER_PORT: [Service MQTT port]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux/cmd/mqtt

# compile the app; make sure to set the proper GOOS value
make mqtt

# set the environment variables and run the service
MF_MANAGER_URL=[Manager service URL] MF_NATS_URL=[NATS instance URL] MF_MQTT_ADAPTER_PORT=[Service MQTT port] app
```

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).
