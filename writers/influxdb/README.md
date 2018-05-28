# InfluxDB writer

InfluxDB writer provides message repository implementation for InfluxDB.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                 | Description                       | Default               |
|--------------------------|-----------------------------------|-----------------------|
| MF_NATS_URL              | NATS instance URL                 | nats://localhost:4222 |
| MF_INFLUX_WRITER_PORT    | Service HTTP port                 | 8180                  |
| MF_INFLUX_WRITER_POINT   | InfluxDB point to write data to   | messages              |
| MF_INFLUX_WRITER_DB_NAME | InfluxDB database name            | mainflux              |
| MF_INFLUX_WRITER_DB_HOST | InfluxDB host                     | localhost             |
| MF_INFLUX_WRITER_DB_PORT | Default port of InfluxDB database | 8086                  |
| MF_INFLUX_WRITER_DB_USER | Default user of InfluxDB database | mainflux              |
| MF_INFLUX_WRITER_DB_PASS | Default password of InfluxDB user | mainflux              |

## Deployment

To start the service, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux


cd $GOPATH/src/github.com/mainflux/mainflux

# compile the influxdb
make influxdb

# copy binary to bin
make install
```

Set the environment variables and run the service
Pass list of env variables in form `VARIABLE_NAME=[value]` separated by space character.
Environment variables are provided in table above. For example:
MF_NATS_URL=nats://localhost:456 MF_INFLUX_WRITER_POINT=point $GOBIN/mainflux-influxdb

## Usage

Starting service will start consuming normalized messages in SenML format.
