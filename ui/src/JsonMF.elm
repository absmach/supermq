-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


module JsonMF exposing (JsonValue(..), emptyString, jsonValueDecoder, jsonValueToString, jsonValueToValue, maybeJsonValueToJsonValue, maybeJsonValueToString, stringToJsonValue, stringToMaybeJsonValue)

import Json.Decode as D
import Json.Encode as E



-- JSONVALUE


type JsonValue
    = ValueObject (List ( String, JsonValue ))
    | ValueArray (List JsonValue)
    | ValueString String
    | ValueFloat Float
    | ValueInt Int
    | ValueBool Bool
    | ValueNull


emptyString : String
emptyString =
    ""


jsonValueDecoder : D.Decoder JsonValue
jsonValueDecoder =
    D.oneOf
        [ D.keyValuePairs (D.lazy (\_ -> jsonValueDecoder)) |> D.map ValueObject
        , D.list (D.lazy (\_ -> jsonValueDecoder)) |> D.map ValueArray
        , D.int |> D.map ValueInt
        , D.float |> D.map ValueFloat
        , D.bool |> D.map ValueBool
        , D.string |> D.map ValueString
        , D.null emptyString |> D.map (\_ -> ValueNull)
        ]


stringToJsonValue : String -> Result D.Error JsonValue
stringToJsonValue jsonString =
    D.decodeString jsonValueDecoder jsonString


jsonValueToValue : JsonValue -> E.Value
jsonValueToValue json =
    case json of
        ValueObject dict ->
            dict
                |> List.map
                    (\( k, v ) ->
                        ( k, jsonValueToValue v )
                    )
                |> E.object

        ValueArray array ->
            array
                |> E.list jsonValueToValue

        ValueString str ->
            E.string str

        ValueFloat number ->
            E.float number

        ValueInt number ->
            E.int number

        ValueBool bool ->
            E.bool bool

        ValueNull ->
            E.null


jsonValueToString : JsonValue -> String
jsonValueToString json =
    json |> jsonValueToValue |> E.encode 0



-- String -> Maybe JsonValue -> JsonValue


stringToMaybeJsonValue : String -> Maybe JsonValue
stringToMaybeJsonValue string =
    case stringToJsonValue string of
        Ok jsonValue ->
            Just jsonValue

        Err err ->
            Just ValueNull


maybeJsonValueToJsonValue : Maybe JsonValue -> JsonValue
maybeJsonValueToJsonValue maybeJsonValue =
    case maybeJsonValue of
        Just jsonValue ->
            jsonValue

        Nothing ->
            ValueNull


maybeJsonValueToString : Maybe JsonValue -> String
maybeJsonValueToString maybeJsonValue =
    jsonValueToString (maybeJsonValueToJsonValue maybeJsonValue)
