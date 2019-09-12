# Starter-kit

Starter-kit service provides a barebones HTTP API for development of a Mainflux
service.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable           | Description                                                  | Default |
|--------------------|--------------------------------------------------------------|---------|
| MF_KIT_LOG_LEVEL   | Log level for starter-kit service (debug, info, warn, error) | error   |
| MF_KIT_HTTP_PORT   | Starter-kit service HTTP port                                | 9021    |
| MF_KIT_SERVER_CERT | Path to server certificate in pem format                     |         |
| MF_KIT_SERVER_KEY  | Path to server key in pem format                             |         |
| MF_JAEGER_URL      | Jaeger server URL                                            |         |
| MF_KIT_SECRET      | Starter-kit service secret                                   | secret  |

## Deployment

The service itself is distributed as Docker container. The following snippet
provides a compose file template that can be used to deploy the service container
locally:

```yaml
version: "2"
services:
  starter-kit:
    image: mainflux/starter-kit:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured HTTP port]
    environment:
      MF_KIT_LOG_LEVEL: [Kit log level]
      MF_KIT_HTTP_PORT: [Service HTTP port]
      MF_KIT_SERVER_CERT: [String path to server cert in pem format]
      MF_KIT_SERVER_KEY: [String path to server key in pem format]
      MF_KIT_SECRET: [Starter-kit service secret]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the kit
make starter-kit

# copy binary to bin
make install

# set the environment variables and run the service
MF_KIT_LOG_LEVEL=[Kit log level] MF_KIT_HTTP_PORT=[Service HTTP port] MF_KIT_SERVER_CERT: [String path to server cert in pem format] MF_KIT_SERVER_KEY: [String path to server key in pem format] MF_KIT_SECRET: [Starter-kit service secret] $GOBIN/mainflux-kit
```

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).

[doc]: http://mainflux.readthedocs.io
