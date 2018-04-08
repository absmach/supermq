# Getting Started
Mainflux is a composition of microservices.

Docker containers are provided for each [Mainflux component](http://mainflux.readthedocs.io/en/latest/#architecture).

## Prerequisites

1. [Docker](https://www.docker.com/community-edition). If you are not familiar with Docker, take a moment to install Docker and work through the Quickstart guide. We'll still be here.
2. [jsonpp](https://jmhodges.github.io/jsonpp/) (optional). Pretty printing json in your terminal.
3. [Mosquitto](https://mosquitto.org/download/) (optional). Mainflux MQTT require the mosquitto cli broker.
4. [Go](https://golang.org/dl/) (optional). Minimum version required is 1.8.

## Installation
To install and deploy the Mainflux system, clone the repo:
```bash
git clone https://github.com/Mainflux/mainflux.git && cd mainflux
```

Then start the Docker composition:
```bash
docker-compose up
```

From sources:

To download all the sources, and place them in appropriate locations (i.e. $GOPATH), use the
[installation script](https://github.com/mainflux/mainflux/blob/master/bin/mainflux-install.sh). Once it completes, the script will provide the
instructions on how to finish the manual installation (i.e. install the required infrastructure).

## Basic Concepts
The Mainflux system operates on four main entities (data structures): Users, Devices, Channels and Applications. Devices and Applications are **clients** of the Mainflux system (i.e. they connect to Mainflux servers and send/receive some messages).

1. `User` represents the real (human) user of the system. User logs into the system and creates some assets - Devices and Applications - that he will own and manage later though Mainflux system. User structure can be seen [here](https://github.com/mainflux/mainflux/blob/master/manager/users.go)

2. `Device` is used to represent any device that connects to Mainflux. It is a generic model
that describes any client device of the system via simple structure like [this](https://github.com/mainflux/mainflux/blob/master/manager/clients.go).

3. `Application` is very similar to the `Device` and is represented by the same `Client` structure (just with different `type` field). Application represents an end-user application that communicates with devices via Mainflux, and can be running somewhere in the cloud, locally on the PC or on the mobile phone - it is usually some UI app that shows dashboards and graphs of the sensor measurements.

3. `Channel` is used to model a communication channel. It is a generic bidirectional message stream representation.
Channels are like MQTT topics - several devices or applications can subscribe or publish on a channel.
All the values that flow through channels are persisted in the database.

Using these four simple representations (Users, Devices, Applications and Channels), it is possible to model complex IoT systems, which mostly include device amangement and IoT messaging tasks.

## SenML
Mainflux uses an [IETF standardized message format called SenML](https://tools.ietf.org/html/draft-ietf-core-senml-08) to exchange IoT messages in the system.

A Mainflux message is a JSON array of SenML hashes. This message is a temperature reading in Celsius:

```
[
     {"n":"urn:dev:ow:10e2073a01080063","u":"Cel","v":23.1}
]
```

SenML permits multiple measurements to be sent in one message. This message contains voltage and current readings:

```
[
     {"bn":"urn:dev:ow:10e2073a0108006;", "bt":1.276020076001e+09, "bu":"A", "bver":5, "n":"voltage","u":"V","v":120.1},
     {"n":"current","t":-5,"v":1.2},
     {"n":"current","t":-4,"v":1.3},
     {"n":"current","t":-3,"v":1.4},
     {"n":"current","t":-2,"v":1.5},
     {"n":"current","t":-1,"v":1.6},
     {"n":"current","v":1.7}
]
```

## Creating User
### Create User Account
Use the Mainflux [manager](https://github.com/mainflux/manager) HTTP/REST API to create user account:
```bash
curl -s -S -i -X POST -H "Content-Type: application/senml+json" localhost:8180/users -d '{"email":"john.doe@email.com", "password":"123"}'
```

> Note that when using official `docker-compose`, all services are behind `NginX` proxy and all traffic is `TLS` encrypted.
You will need following command in this case:
```
curl -s -S -i --cacert docker/ssl/certs/mainflux-server.crt --insecure -X POST -H "Content-Type: application/senml+json" https://localhost/tokens -d '{"email":"john.doe@email.com", "password":"123"}'
```

### Create User Token (Authorization Key)
In order for this user to be capable to authenticate to the system, we must create for him an authorization key - i.e. user token:
```bash
curl -s -S -i -X POST -H "Content-Type: application/senml+json" localhost:8180/tokens -d '{"email":"john.doe@email.com", "password":"123"}'
```

Response:
```bash
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDQ0MzE5MDUsImlhdCI6MTUwNDM5NTkwNSwiaXNzIjoibWFpbmZsdXgiLCJzdWIiOiJqb2huLmRvZUBlbWFpbC5jb20ifQ.BhcQCuSKy-XLG9tYa5wp8Suue8sOQ7pAgRPiExVzgwo"}
```

## System Provisioning
Configuring your Mainflux system begins with a provisioning phase. Use the Mainflux [manager](https://github.com/mainflux/mainflux/tree/master/manager) HTTP/REST API to provision Devices, Applications and Channels.

### Provisioning Devices
Devices are provisioned by executing a HTTP request `POST /clients`, with a `"type":"device"` specified in the payload JSON.
Note that you will need `user_auth_key`, i.e. user token, in order to provision clients (devices and applications) that belong to this particular user.

```bash
curl -s -S -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/clients -d '{"type":"device", "name":"weio"}'
```

Notice the UUID of the device in the Location header of the HTTP response:
```
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Location: /clients/8293b8fa-9039-11e7-b6e2-080027b77be6
Date: Sat, 02 Sep 2017 23:50:33 GMT
Content-Length: 0
```

It is important to note that the Mainflux system generates a UUID for each Device, Application or Channel it creates.
This ID is later used to authenticate each Device on the system.

### Provisioning Applications
Applications are also Mainflux clients, as are devices. They are created in a similar fashion:

```bash
curl -s -S -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/clients -d '{"type":"app", "name":"myapp"}'
```

You'll find UUID of newly created app in HTTP response `Location` header:
```
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Location: /clients/75a31f6f-90ad-11e7-9cef-080027b77be6
Date: Sun, 03 Sep 2017 13:40:33 GMT
Content-Length: 0
```

### Retrieving Clients (Devices and Apps)
In order to check of provisioning went fine, you cold look what entities were written in the Cassandra database usin `GET /devices`:
```bash
curl -s -S -i --noproxy localhost -X GET -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/clients
```
You should obtain response similar to this:
```json
{
  "clients": [
    {
      "id": "8293b8fa-9039-11e7-b6e2-080027b77be6",
      "type": "device",
      "name": "weio",
      "key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI"
    },
    {
      "id": "75a31f6f-90ad-11e7-9cef-080027b77be6",
      "type": "app",
      "name": "myapp",
      "key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQ0NDYwMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiNzVhMzFmNmYtOTBhZC0xMWU3LTljZWYtMDgwMDI3Yjc3YmU2In0.MEJxXI37LOGcnm50ap8DsPVQa56mg4nQ0W7ZcDE785k"
    }
  ]
}
```

### Provisioning Channels
Channels are provisioned by executing a HTTP request `POST /channels`:
```bash
curl -s -S -i --noproxy localhost -X POST -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/channels -d '{"name":"mychan"}'
```

Again, UUID of newly created resource is returned in `Location` HTTP header:
```bash
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Location: /channels/7209d9b8-90af-11e7-9cf0-080027b77be6
Date: Sun, 03 Sep 2017 13:54:46 GMT
Content-Length: 0
```

### Grouping Clients by Connecting Them on Channels
Channel can be observed as a communication group of clients, i.e. and entity that is used to isolate several clients (apps and devices) in the same (communication) group. Only clients that are connected to the channel can send and receive messages from this channel. Oter clients are not permited to communicate over this channel.

Only user, who is the owner of the channel and of the clients, can connect the clients to a channel (which is equivalent to putting the clients in the same administrative group, with the persmissions to exchange messages within the group).

Connecting clients to the channel is done by adding them in the `connected` JSON payload field during the channel creation, or by updating the existing channel, like this:
```bash
curl -s -S -i --noproxy localhost -X PUT -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/channels/7209d9b8-90af-11e7-9cf0-080027b77be6/clients/8293b8fa-9039-11e7-b6e2-080027b77be6
```

Then you can observe that clients you connected are recorded in the channel's `connected` list:
```
curl -s -S -i --noproxy localhost -X GET -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/channels/7209d9b8-90af-11e7-9cf0-080027b77be6

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sun, 03 Sep 2017 14:05:19 GMT
Content-Length: 154

{
  "id": "7209d9b8-90af-11e7-9cf0-080027b77be6",
  "name": "mychan",
  "connected": [
    {
      "id": "75a31f6f-90ad-11e7-9cef-080027b77be6",
      "type": "device",
      "name": "weio",
      "key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI"
    }
  ]
}
```

## Publishing Values (on a Channel)
Once a Channel is provisioned and client is connected to the channel, it can start to publish messages on the channel.

This can be done via several protocols:

- HTTP
- MQTT
- WebSockets (MQTT over WebSockets)
- CoAP

### HTTP

Publishing and retrieving messages (values) of one particular channel is done via `POST` or `GET` on an API endpoint `/channels/:channel_id/messages` of [http-adapter](https://github.com/mainflux/mainflux/tree/master/http) service:

You will need `client_token`, i.e. auth key of the client (device or application) that tries to publish the message on the channel. You can obtain this key by observing client entity in the database:
```
curl -s -S -i --noproxy localhost -X GET -H "Content-Type: application/senml+json" -H "Authorization: <user_token>" localhost:8180/clients/8293b8fa-9039-11e7-b6e2-080027b77be6

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sun, 03 Sep 2017 16:12:23 GMT
Content-Length: 273

{
  "id": "8293b8fa-9039-11e7-b6e2-080027b77be6",
  "type": "device",
  "name": "weio",
  "key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI"
}
```

Sneding message on a channel:
```bash
curl -s -S -i -X POST -H "Content-Type: application/senml+json" -H "Authorization: <client_token>" localhost:7070/channels/7209d9b8-90af-11e7-9cf0-080027b77be6/messages -d '[{"bn":"some-base-name:","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]'
```

You can now observe contents in the Cassandra database to see written messages:
```bash
cqlsh> SELECT * FROM message_writer.messages_by_channel;

 channel                              | id                                   | bn              | bs | bt        | bu | bv | bver | l | n       | protocol | publisher                                                                                                                                                                                    | s | t  | u | ut | v     | vb    | vd | vs
--------------------------------------+--------------------------------------+-----------------+----+-----------+----+----+------+---+---------+----------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+---+----+---+----+-------+-------+----+----
 7209d9b8-90af-11e7-9cf0-080027b77be6 | fa22b310-90c3-11e7-a258-7584557e38c9 | some-base-name: |  0 | 1.276e+09 |  A |  0 |    5 |   | voltage |     HTTP | eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI | 0 |  0 | V |  0 | 120.1 | False |    |
 7209d9b8-90af-11e7-9cf0-080027b77be6 | fa252410-90c3-11e7-a258-7584557e38c9 |                 |  0 |         0 |    |  0 |    0 |   | current |     HTTP | eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI | 0 | -5 |   |  0 |   1.2 | False |    |
 7209d9b8-90af-11e7-9cf0-080027b77be6 | fa260e70-90c3-11e7-a258-7584557e38c9 |                 |  0 |         0 |    |  0 |    0 |   | current |     HTTP | eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MDQzOTYyMzMsImlzcyI6Im1haW5mbHV4Iiwic3ViIjoiODI5M2I4ZmEtOTAzOS0xMWU3LWI2ZTItMDgwMDI3Yjc3YmU2In0.3OX3kccV2Beh7C87GAB_FPJ6SKQeJVBTUVdQDa39TzI | 0 | -4 |   |  0 |   1.3 | False |    |

(3 rows)
```

### MQTT
Mainflux is acting as a seamless multi-protocol bridge. If you were subscribed to an MQTT topic `mainflux/channels/:channel_id/messages/<content_type>` you would get the message published via HTTP POST on `mainflux/channels/:channel_id/messages` with `Content-Type: application/senml+json` header.

Here it is important to note that we could not pass `Content-Type` in some kind of a header, as in the case of HTTP, so we use topic itself to denote the type of content that is sent over the channel.

Mainflux currently supports `application/senml+json` and `application/octet-stream` (for sending custom binary blobs), but in the future a support for `application/senml+cbor` will be added.

These topics are denoted as following:

- `mainflux/channels/5c912c4e-e37b-4ba6-8f4b-373c7ecfeaa9/messages/senml-json`
- `mainflux/channels/5c912c4e-e37b-4ba6-8f4b-373c7ecfeaa9/messages/senml-cbor`
- `mainflux/channels/5c912c4e-e37b-4ba6-8f4b-373c7ecfeaa9/messages/octet-stream`

Note that `+` is a wildcard character for MQTT topics, so we use `senml-json` to denote `application/senml+json` content type.

For security purposes (auth), it is required that each client presents it's token as an  MQTT "clientID" (part of MQTT standard).

Example:

```
mosquitto_sub -u <client_id> -P <client_token> -t mainflux/channels/7209d9b8-90af-11e7-9cf0-080027b77be6/messages/senml-json

[{"bn":"8293b8fa-9039-11e7-b6e2-080027b77be6","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]
```

Publishing via MQTT is done in the similar way:
```
mosquitto_pub -u <client_id> -P <client_token> -t mainflux/channels/7209d9b8-90af-11e7-9cf0-080027b77be6/messages/senml-json -m '[{"bn":"8293b8fa-9039-11e7-b6e2-080027b77be6","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]'
```

### MQTT over SSL/TLS

Example with `letsencrpyt`:

On your Mac, open `Keychain Access` -> Search "DST Root CA X3" -> Export "DST Root CA X3" as `DST_Root_CA_X3.pem` or `DST_Root_CA_X3.crt` and save it.

Subscribing via MQTT over SSL/TLS:

```shell
mosquitto_sub -h mqtt.yourflux.com -p 8883 -u <client_id> -P <client_token> -t mainflux/channels/<channel_id>/messages/senml-json -d -v --cafile certs/DST_Root_CA_X3.pem --cert certs/mqtt.yourflux.com/cert.pem --key certs/mqtt.yourflux.com/key.pem
```

Publishing via MQTT over SSL/TLS:

```shell
mosquitto_pub -h mqtt.yourflux.com -p 8883 -u <client_id> -P <client_token> -t mainflux/channels/<channel_id>/messages/senml-json -d -m '[{"bn":"15c33152-db54-11e7-bb8f-0242ac110005","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]' --cafile certs/DST_Root_CA_X3.pem --cert certs/mqtt.yourflux.com/cert.pem --key certs/mqtt.yourflux.com/key.pem
```


### Websockets
Mainflux also supports Websockets with MQTT over WS.
Every modern browser or any device is now a potential full-fledged MQTT client.
With publish/subscribe, quality of service and retain messages, clients like web apps can take full advantage
of highly scalable messaging with a very low bandwidth footprint.

Similar to MQTT, the WebSockets API supports publish and subscribe to any channel/topic on same end point as MQTT  `mainflux/channels/:channel_id/messages/senml-json`.

[Here](https://github.com/mainflux/mainflux-mqtt/blob/master/examples/paho-js-client/index.html) you will
find an example of web client implementation using the Eclipse [Paho javascript library](https://eclipse.org/paho/clients/js/)
