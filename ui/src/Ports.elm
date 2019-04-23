-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


port module Ports exposing (Websocket, channelIdFromUrl, connectWebsocket, disconnectWebsocket, queryWebsocket, retrieveWebsocket, websocketDecoder, websocketIn, websocketOut, websocketState)

import Json.Decode as D
import Json.Encode as E


port websocketIn : (String -> msg) -> Sub msg


port websocketState : (String -> msg) -> Sub msg


port websocketOut : E.Value -> Cmd msg


port connectWebsocket : E.Value -> Cmd msg


port disconnectWebsocket : E.Value -> Cmd msg


port queryWebsocket : E.Value -> Cmd msg


port retrieveWebsocket : (E.Value -> msg) -> Sub msg



-- JSON


type alias Websocket =
    { url : String
    , readyState : Int
    }


websocketDecoder : D.Decoder Websocket
websocketDecoder =
    D.map2 Websocket
        (D.field "url" D.string)
        (D.field "readyState" D.int)


channelIdFromUrl : String -> String
channelIdFromUrl url =
    let
        start =
            String.length "wss://localhost/ws/channels/"

        end =
            String.length "wss://localhost/ws/channels/"
                + String.length "0522c54b-5b00-4aab-a2b0-6e3e54320995"
    in
    String.slice start end url
