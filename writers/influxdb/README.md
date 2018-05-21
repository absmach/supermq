# InfluxDB writer

InfluxDB writer provides message repository implementation for InfluxDB.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                   | Description                       | Default               |
|----------------------------|-----------------------------------|-----------------------|
| MF_INFLUXDB_DB_WRITER_PORT | Service HTTP port                 | 8180                  |
| MF_NATS_URL                | NATS instance URL                 | nats://localhost:4222 |
| MF_INFLUXDB_DB_POINT_NAME  | InfluxDB point to write data to   | messages              |
| MF_INFLUXDB_DB_NAME        | InfluxDB database name            | mainflux              |
| MF_INFLUXDB_DB_HOST        | InfluxDB host                     | localhost             |
| MF_INFLUXDB_DB_PORT        | Default port of InfluxDB database | 8086                  |
| MF_INFLUXDB_DB_USER        | Default user of InfluxDB database | mainflux              |
| MF_INFLUXDB_DB_PASS        | Default password of InfluxDB user | mainflux              |

## Deployment

To start the service, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux


cd $GOPATH/src/github.com/mainflux/mainflux

# compile the http
CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH GOARM=$GOARM go build -ldflags "-s -w" -o ./build/mainflux-influxdb cmd/influxdb/main.go

# copy binary to bin
cp ./build/* $GOBIN
```

Set the environment variables and run the service
Pass list of env variables in form `VARIABLE_NAME=[value]` separated by space character.
Env variables are provided in table above. For example:
MF_NATS_URL=nats://localhost:456 MF_DB_USER=user $GOBIN/mainflux-influxdb

## Usage

Starting service will start consuming messages normalized messages in SenML format.
