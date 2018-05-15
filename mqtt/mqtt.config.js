'use strict';

// Service configuration
module.exports = {
    mqtt_port: process.env.MF_MQTT_ADAPTER_PORT || 1883,
    ws_port: process.env.WS_PORT || 8880,
    // NATS broker URL
    nats_url: process.env.MF_NATS_URL || 'nats://localhost:4222',
    // Auth service URL
    auth_url: process.env.MF_THINGS_URL || 'localhost:8181',
};
