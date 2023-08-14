# Mainflux Modbus Adapter

The Mainflux Modbus Adapter service is responsible for reading and writing data to Modbus sensors using various protocols such as TCP and RTU/ASCII. It serves as an interface between Mainflux and Modbus devices, allowing you to easily integrate Modbus devices into your IoT ecosystem.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                         | Description                                         | Default                        |
| -------------------------------- | --------------------------------------------------- | ------------------------------ |
| MF_MODBUS_ADAPTER_LOG_LEVEL        | Service log level                                   | info                           |
| MF_BROKER_URL                    | Message broker instance URL                         | nats://localhost:4222          |
| MF_JAEGER_URL                    | Jaeger server URL                                   | http://jaeger:14268/api/traces |
| MF_SEND_TELEMETRY                | Send telemetry to mainflux call home server         | true                           |
| MF_MODBUS_ADAPTER_INSTANCE_ID      | Modbus adapter instance ID                            |                                |

## Deployment

The service itself is distributed as Docker container. Check the [`modbus-adapter`](https://github.com/mainflux/mainflux/blob/master/docker/docker-compose.yml#L273-L291) service section in
docker-compose to see how service is deployed.

Running this service outside of container requires working instance of the message broker service.
To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/mainflux/mainflux

cd mainflux

# compile the http
make modbus

# copy binary to bin
make install

# set the environment variables and run the service
MF_MODBUS_ADAPTER_LOG_LEVEL=[Service log level] \
MF_BROKER_URL=[Message broker instance URL] \
MF_JAEGER_URL=[Jaeger server URL] \
MF_SEND_TELEMETRY=[Send telemetry to mainflux call home server] \
MF_MODBUS_ADAPTER_INSTANCE_ID=[CoAP adapter instance ID] \
$GOBIN/mainflux-coap
```

## Usage

The Mainflux Modbus Adapter service interacts with Modbus sensors by subscribing to specific channels for reading and writing Modbus values. It utilizes the Mainflux messaging system and follows a specific payload structure for configuration.

### Reading Values

To start reading values, you need to publish a message using mainflux messaging adapters such as http, coap, mqtt etc to the channel `channels/<channel_id>/messages/modbus/read/<modbus_protocol>/<modbus_data_point>`.

The supported modbus protocols include:

- TCP
- RTU

The supported data points include:

- coil
- h_register
- i_register
- register
- discrete
- fifo

The payload of the message is structured as follows:

```json
{
    "options": {
        "address": 123,
        "quantity": 2,
    },
    "config": {}
}

```

The config can be either TCP or RTU and has the following structure:

```go
type RTUHandlerOptions struct {
	Address     string             `json:"address,omitempty"`
	BaudRate    int                `json:"baud_rate,omitempty"`
	Config      serial.Config      `json:"config,omitempty"`
	DataBits    int                `json:"data_bits,omitempty"`
	IdleTimeout time.Duration      `json:"idle_timeout,omitempty"`
	Logger      *log.Logger        `json:"-"`
	Parity      string             `json:"parity,omitempty"`
	RS485       serial.RS485Config `json:"rs485,omitempty"`
	SlaveId     byte               `json:"slave_id,omitempty"`
	StopBits    int                `json:"stop_bits,omitempty"`
	Timeout     time.Duration      `json:"timeout,omitempty"`
}
type TCPHandlerOptions struct {
	Address     string        `json:"address"`
	IdleTimeout time.Duration `json:"idle_time"`
	Logger      *log.Logger   `json:"-"`
	SlaveId     byte          `json:"slave_id,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}
```

The results of the readings are published on `channels/<channel_id>/messages/modbus/res`

### Writing Values

To start reading values, you need to publish a message using mainflux messaging adapters such as http, coap, mqtt etc to the channel `channels/<channel_id>/messages/modbus/write/<modbus_protocol>/<modbus_data_point>`.

The payload of the message is structured as follows:

```json
{
    "options": {
        "address": 123,
        "quantity": 2,
        "value": {}
    },
    "config": {}
}
```

The value field can be either `uint16` or `[]byte`.

The results of the readings are published on `channels/<channel_id>/messages/modbus/res`
