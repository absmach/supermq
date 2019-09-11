# Starter-kit

Starter-kit service provides a barebones HTTP API for development of a Mainflux
service.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                      | Description                                                  | Default |
|-------------------------------|--------------------------------------------------------------|---------|
| MF_KIT_LOG_LEVEL      | Log level for starter-kit service (debug, info, warn, error) | error   |
| MF_KIT_HTTP_PORT      | Starter-kit service HTTP port                                | 8180    |
| MF_KIT_AUTH_HTTP_PORT | Starter-kit service auth HTTP port                           | 8989    |
| MF_KIT_AUTH_GRPC_PORT | Starter-kit service auth gRPC port                           | 8181    |

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
      MF_KIT_LOG_LEVEL: [Things log level]
      MF_KIT_HTTP_PORT: [Service HTTP port]
      MF_KIT_AUTH_HTTP_PORT: [Service auth HTTP port]
      MF_KIT_AUTH_GRPC_PORT: [Service auth gRPC port]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the things
make starter-kit

# copy binary to bin
make install

# set the environment variables and run the service
MF_KIT_LOG_LEVEL=[Things log level] MF_KIT_HTTP_PORT=[Service HTTP port] MF_KIT_AUTH_HTTP_PORT=[Service auth HTTP port] MF_KIT_AUTH_GRPC_PORT=[Service auth gRPC port] $GOBIN/mainflux-things
```

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).

[doc]: http://mainflux.readthedocs.io
