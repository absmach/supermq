-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


port module Ports exposing (connectWebsocket, disconnectWebsocket, websocketIn, websocketOut, websocketState)

import Json.Encode as E


port websocketIn : (String -> msg) -> Sub msg


port websocketState : (String -> msg) -> Sub msg


port websocketOut : E.Value -> Cmd msg


port connectWebsocket : E.Value -> Cmd msg


port disconnectWebsocket : E.Value -> Cmd msg
