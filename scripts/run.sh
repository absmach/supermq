#!/bin/bash
#
# Copyright (c) 2018
# Mainflux
#
# SPDX-License-Identifier: Apache-2.0
#

###
# Runs all Mainflux microservices (must be previously built and installed).
#
# Expects that PostgreSQL and needed messaging DB are alredy running.
#
###


# Kill all mainflux-* stuff
function cleanup {
	pkill mainflux
    pkill nats
}

###
# NATS
###
gnatsd &

###
# Users
###
mainflux-users &

###
# Things
###
MF_THINGS_HTTP_PORT=8182 MF_THINGS_GRPC_PORT=8183 mainflux-things &


###
# HTTP
###
MF_HTTP_ADAPTER_PORT=8185 MF_THINGS_URL=localhost:8183 mainflux-http &

###
# WS
###
MF_WS_ADAPTER_PORT=8186 MF_THINGS_URL=localhost:8183 mainflux-ws &

###
# MQTT
###


###
# CoAP
###


trap cleanup EXIT

while : ; do sleep 1 ; done