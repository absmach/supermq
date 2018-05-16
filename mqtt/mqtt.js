/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0 license.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

'use strict';

var http = require('http');
var websocket = require('websocket-stream');
var net = require('net');
var aedes = require('aedes')();
var logging = require('aedes-logging');
var request = require('request');
var util = require('util');

var config = require('./mqtt.config');
var nats = require('nats').connect(config.nats_url);

var protobuf = require('protocol-buffers');
var grpc = require('grpc');
var fs = require('fs');

// pass a proto file as a buffer/string or pass a parsed protobuf-schema object
var message = protobuf(fs.readFileSync('../message.proto'));
var things = grpc.load("../internal.proto").mainflux;

var thingsClient = new things.ThingsService(config.auth_url, grpc.credentials.createInsecure());

var servers = [
    startWs(),
    startMqtt()
];

logging({
    instance: aedes,
    servers: servers
});

/**
 * WebSocket
 */
function startWs() {
    var server = http.createServer();
    websocket.createServer({server: server}, aedes.handle);
    server.listen(config.ws_port);
    return server;
}

/**
 * MQTT
 */
function startMqtt() {
    return net.createServer(aedes.handle).listen(config.mqtt_port);
}

/**
 * NATS
 */
nats.subscribe('channel.*', function (msg) {
    var m = message.RawMessage.decode(Buffer.from(msg));

    // Parse and adjust content-type
    if (m.ContentType === "application/senml+json") {
        m.ContentType = "senml-json";
    }

    var packet = {
        cmd: 'publish',
        qos: 2,
        topic: 'channels/' + m.Channel + '/messages/' + m.ContentType,
        payload: m.Payload,
        retain: false
    };
    console.log("Received message from NATS");
    console.log(packet);

    aedes.publish(packet);
});

/**
 * Hooks
 */
// AuthZ PUB
aedes.authorizePublish = function (client, packet, callback) {
    // Topics are in the form `channels/<channel_id>/messages/senml-json`
    var channel = packet.topic.split('/')[1];

    thingsClient.CanAccess({
        token: client.password,
        chanID: channel
    }, function (err, res) {
        if (!err) {
            console.log('Publish authorized OK');

            /**
             * We must publish on NATS here, because on_publish() is also called
             * when we receive message from NATS from other adapters (in nats.subscribe()),
             * so we must avoid re-publishing on NATS what came from other adapters
             */
            var msg = {
                Publisher: client.id,
                Channel: channel,
                Protocol: 'mqtt',
                ContentType: packet.topic.split('/')[3],
                Payload: packet.payload
            };
            var rawMsg = message.RawMessage.encode(msg);

            console.log("msg:", msg);
            console.log("packet:", util.inspect(packet, false, null));

            // Pub on NATS
            nats.publish('channel.' + channel, rawMsg);
        } else {
            console.log('Publish not authorized');
            callback(4); // Bad username or password
        }
    });
};

// AuthZ SUB
aedes.authorizeSubscribe = function (client, packet, callback) {
    // Topics are in the form `channels/<channel_id>/messages/senml-json`
    var channel = packet.topic.split('/')[1];
    
    thingsClient.canAccess({
        token: client.password,
        chanID: channel
    }, function (err, res) {
        if (!err) {
            console.log('Subscribe authorized OK');
            callback(null, packet);
        } else {
            console.log('Subscribe not authorized');
            callback(4, packet); // Bad username or password
        }
    });
};

// AuthX
aedes.authenticate = function (client, username, password, callback) {
    client.id = username.toString() || "";
    client.password = password.toString() || "";
    callback(null, true);
};

/**
 * Handlers
 */
aedes.on('clientDisconnect', function (client) {
    console.log('client disconnect', client.id);
    // Remove client password
    client.password = null;
});

aedes.on('clientError', function (client, err) {
  console.log('client error', client.id, err.message, err.stack);
});

aedes.on('connectionError', function (client, err) {
  console.log('client error', client, err.message, err.stack);
});
