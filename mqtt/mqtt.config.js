'use strict';

// Service configuration
module.exports = {
    mqtt_port: process.env.MF_MQTT_ADAPTER_PORT || 1883,
    ws_port: process.env.MF_WS_PORT || 8880,
    nats_url: process.env.MF_NATS_URL || 'nats://localhost:4222',
    auth_url: process.env.MF_THINGS_URL || 'localhost:8181',
};
