# Mainflux message writer

Mainflux message writer consumes channel events published on message broker,
and stores them into the database.

## Configuration

The service requires only one configuration parameter - Consul agent URL. It is
expected to be set through the `CONSUL_ADDR` environment variable.

For service to run properly, Consul cluster must contain the following keys:

| Key            | Description                                         |
|----------------|-----------------------------------------------------|
| cassandra      | comma-separated contact points in Cassandra cluster |
| nats           | NATS instance URL                                   |

## Deployment

Before proceeding to deployment, make sure to check out the [Apache Cassandra 3.0.x
documentation][www:cassandra]. Developers are advised to get acquainted with
basic architectural concepts, data modeling techniques and deployment strategies.

> Prior to deploying the service, make sure to set up the database and create
the keyspace that will be used by the service.

The service itself is distributed as Docker container. The following snippet
provides a compose file template that can be used to deploy the service container
locally:

```yaml
version: "2"
services:
  manager:
    image: mainflux/message-writer:[version]
    container_name: [instance name]
    environment:
      MESSAGE_WRITER_DB_CLUSTER: [comma-separated Cassandra endpoints]
      MESSAGE_WRITER_DB_KEYSPACE: [name of Cassandra keyspace]
      MESSAGE_WRITER_NATS_URL: [NATS instance URL]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/message-writer

cd $GOPATH/github.com/mainflux/message-writer/cmd

# compile the app; make sure to set the proper GOOS value
CGO_ENABLED=0 GOOS=[platform identifier] go build -ldflags "-s" -a -installsuffix cgo -o app

# set the environment variables and run the service
MESSAGE_WRITER_DB_CLUSTER=[comma-separated Cassandra endpoints] MESSAGE_WRITER_DB_KEYSPACE=[name of Cassandra keyspace] MESSAGE_WRITER_NATS_URL=[NATS instance URL] app
```

[www:cassandra]: http://docs.datastax.com
