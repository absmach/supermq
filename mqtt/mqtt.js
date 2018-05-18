'use strict';

var http = require('http');
var net = require('net');
var aedes = require('aedes')();
var logging = require('aedes-logging');
var protobuf = require('protocol-buffers');
var websocket = require('websocket-stream');
var grpc = require('grpc');
var fs = require('fs');
var bunyan = require('bunyan');

var config = require('./mqtt.config');
var nats = require('nats').connect(config.nats_url);

// pass a proto file as a buffer/string or pass a parsed protobuf-schema object
var logger = bunyan.createLogger({name: "mqtt"}),
    message = protobuf(fs.readFileSync('../message.proto')),
    thingsSchema = grpc.load("../internal.proto").mainflux,
    things = new thingsSchema.ThingsService(config.auth_url, grpc.credentials.createInsecure()),
    servers = [
        startMqtt(),
        startWs()
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
    var m = message.RawMessage.decode(Buffer.from(msg)),
        packet = {
            cmd: 'publish',
            qos: 2,
            topic: 'channels/' + m.Channel + '/messages',
            payload: m.Payload,
            retain: false
        };

    aedes.publish(packet);
});

/**
 * Hooks
 */
// AuthZ PUB
aedes.authorizePublish = function (client, packet, callback) {
    // Topics are in the form `channels/<channel_id>/messages`
    var channel = packet.topic.split('/')[1];

    things.CanAccess({
        token: client.password,
        chanID: channel
    }, function (err, res) {
        if (!err) {
            logger.info('authorized publish');
            /**
             * We must publish on NATS here, because on_publish() is also called
             * when we receive message from NATS from other adapters (in nats.subscribe()),
             * so we must avoid re-publishing on NATS what came from other adapters
             */
            var rawMsg = message.RawMessage.encode({
                Publisher: client.id,
                Channel: channel,
                Protocol: 'mqtt',
                Payload: packet.payload
            });

            // Pub on NATS
            nats.publish('channel.' + channel, rawMsg);
        } else {
            logger.warn("unauthorized publish: %s", err.message);
            callback(4); // Bad username or password
        }
    });
};

// AuthZ SUB
aedes.authorizeSubscribe = function (client, packet, callback) {
    // Topics are in the form `channels/<channel_id>/messages`
    var channel = packet.topic.split('/')[1];
    
    things.canAccess({
        token: client.password,
        chanID: channel
    }, function (err, res) {
        if (!err) {
            logger.info('authorized subscribe');
            callback(null, packet);
        } else {
            logger.warn('unauthorizerd subscribe: %s', err);
            callback(4, packet); // Bad username or password
        }
    });
};

// AuthX
aedes.authenticate = function (client, username, password, callback) {
    var pass = password || "";
    pass = pass.toString() || "";
    things.identify({value: pass}, function(err, res) {
        if (!err) {
            client.id = res.value.toString() || "";
            client.password = pass;
            callback(null, true);
        } else {
            logger.warn('failed to authenticate client with key %s', password);
            callback(err, false);
        }
    });
};

/**
 * Handlers
 */
aedes.on('clientDisconnect', function (client) {
    logger.info('disconnect client %s', client.id);
    // Remove client password
    client.password = null;
    
});

aedes.on('clientError', function (client, err) {
  logger.warn('client error: client: %s, error: %s, stack: %s', client.id, err.message, err.stack);
});

aedes.on('connectionError', function (client, err) {
  logger.warn('client error: client: %s, error: %s, stack: %s', client.id, err.message, err.stack);
});
