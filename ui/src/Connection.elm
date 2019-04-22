-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


port module Connection exposing (Model, Msg(..), initial, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Bootstrap.Utilities.Spacing as Spacing
import Channel
import Debug exposing (log)
import Error
import Helpers exposing (faIcons, fontAwesome)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import HttpMF exposing (paths)
import Json.Encode as E
import List.Extra
import Ports exposing (..)
import Thing
import Url.Builder as B


type alias Model =
    { response : String
    , things : Thing.Model
    , channels : Channel.Model
    , checkedThingsIds : List String
    , checkedThingsKeys : List String
    , checkedChannelsIds : List String
    , websocketData : String
    }


initial : Model
initial =
    { response = ""
    , things = Thing.initial
    , channels = Channel.initial
    , checkedThingsIds = []
    , checkedThingsKeys = []
    , checkedChannelsIds = []
    , websocketData = ""
    }


type Msg
    = Connect
    | Disconnect
    | Listen
    | WebsocketData String
    | Stop
    | ThingMsg Thing.Msg
    | ChannelMsg Channel.Msg
    | GotResponse (Result Http.Error String)
    | CheckThing ( String, String )
    | CheckChannel String


resetChecked : Model -> Model
resetChecked model =
    { model | checkedThingsIds = [], checkedThingsKeys = [], checkedChannelsIds = [] }


isEmptyChecked : Model -> Bool
isEmptyChecked model =
    List.isEmpty model.checkedThingsIds || List.isEmpty model.checkedChannelsIds


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        Connect ->
            if isEmptyChecked model then
                ( model, Cmd.none )

            else
                ( resetChecked model
                , Cmd.batch (connect model.checkedThingsIds model.checkedChannelsIds "PUT" token)
                )

        Disconnect ->
            if isEmptyChecked model then
                ( model, Cmd.none )

            else
                ( resetChecked model
                , Cmd.batch (connect model.checkedThingsIds model.checkedChannelsIds "DELETE" token)
                )

        Listen ->
            if isEmptyChecked model then
                ( model, Cmd.none )

            else
                ( resetChecked model
                , Cmd.batch
                    (connect model.checkedThingsIds model.checkedChannelsIds "PUT" token
                        ++ listen model.checkedThingsKeys model.checkedChannelsIds
                    )
                )

        Stop ->
            if isEmptyChecked model then
                ( resetChecked model, Cmd.none )

            else
                ( resetChecked model, Cmd.batch (stop model.checkedThingsKeys model.checkedChannelsIds) )

        WebsocketData data ->
            ( { model | websocketData = data }, Cmd.none )

        GotResponse result ->
            case result of
                Ok statusCode ->
                    ( { model | response = statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        ThingMsg subMsg ->
            let
                ( updatedThing, thingCmd ) =
                    Thing.update subMsg model.things token
            in
            ( { model | things = updatedThing }, Cmd.map ThingMsg thingCmd )

        ChannelMsg subMsg ->
            let
                ( updatedChannel, channelCmd ) =
                    Channel.update subMsg model.channels token
            in
            ( { model | channels = updatedChannel }, Cmd.map ChannelMsg channelCmd )

        CheckThing thing ->
            ( { model
                | checkedThingsIds = Helpers.checkEntity (Tuple.first thing) model.checkedThingsIds
                , checkedThingsKeys = Helpers.checkEntity (Tuple.second thing) model.checkedThingsIds
              }
            , Cmd.none
            )

        CheckChannel id ->
            ( { model | checkedChannelsIds = Helpers.checkEntity id model.checkedChannelsIds }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ websocketState WebsocketData
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                (Helpers.appendIf (model.things.things.total > model.things.limit)
                    [ Helpers.genCardConfig faIcons.things "Things" (genThingRows model.checkedThingsIds model.things.things.list) ]
                    (Html.map ThingMsg (Helpers.genPagination model.things.things.total (Helpers.offsetToPage model.things.offset model.things.limit) Thing.SubmitPage))
                )
            , Grid.col []
                (Helpers.appendIf (model.channels.channels.total > model.channels.limit)
                    [ Helpers.genCardConfig faIcons.channels "Channels" (genChannelRows model.checkedChannelsIds model.channels.channels.list) ]
                    (Html.map ChannelMsg (Helpers.genPagination model.channels.channels.total (Helpers.offsetToPage model.channels.offset model.channels.limit) Channel.SubmitPage))
                )
            ]
        , Grid.row []
            [ Grid.col []
                [ Form.form []
                    [ Button.button [ Button.success, Button.attrs [ Spacing.ml1 ], Button.onClick Connect ] [ text "Connect" ]
                    , Button.button [ Button.danger, Button.attrs [ Spacing.ml1 ], Button.onClick Disconnect ] [ text "Disconnect" ]
                    ]
                ]
            , Grid.col []
                [ Form.form []
                    [ Button.button [ Button.success, Button.attrs [ Spacing.ml1 ], Button.onClick Listen ] [ text "Listen" ]
                    , Button.button [ Button.danger, Button.attrs [ Spacing.ml1 ], Button.onClick Stop ] [ text "Stop" ]
                    ]
                ]
            ]
        , Helpers.response model.response
        , Helpers.response model.websocketData
        ]


genThingRows : List String -> List Thing.Thing -> List (Table.Row Msg)
genThingRows checkedThingsIds things =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ text (" " ++ Helpers.parseString thing.name) ]
                , Table.td [] [ text thing.id ]
                , Table.td [] [ input [ type_ "checkbox", onClick (CheckThing ( thing.id, thing.key )), checked (Helpers.isChecked thing.id checkedThingsIds) ] [] ]
                ]
        )
        things


genChannelRows : List String -> List Channel.Channel -> List (Table.Row Msg)
genChannelRows checkedChannelsIds channels =
    List.map
        (\channel ->
            Table.tr []
                [ Table.td [] [ text (" " ++ Helpers.parseString channel.name) ]
                , Table.td [] [ text channel.id ]
                , Table.td [] [ input [ type_ "checkbox", onClick (CheckChannel channel.id), checked (Helpers.isChecked channel.id checkedChannelsIds) ] [] ]
                ]
        )
        channels



-- HTTP


connect : List String -> List String -> String -> String -> List (Cmd Msg)
connect checkedThingsIds checkedChannelsIds method token =
    List.foldr (++)
        []
        (List.map
            (\thingid ->
                List.map
                    (\channelid ->
                        HttpMF.request
                            (B.relative [ paths.channels, channelid, paths.things, thingid ] [])
                            method
                            token
                            Http.emptyBody
                            GotResponse
                    )
                    checkedChannelsIds
            )
            checkedThingsIds
        )


listen : List String -> List String -> List (Cmd Msg)
listen checkedThingsKeys checkedChannelsIds =
    List.foldr (++)
        []
        (List.map
            (\thingkey ->
                List.map
                    (\channelid ->
                        connectWebsocket <|
                            E.object
                                [ ( "channelid", E.string channelid )
                                , ( "thingkey", E.string thingkey )
                                ]
                    )
                    checkedChannelsIds
            )
            checkedThingsKeys
        )


stop : List String -> List String -> List (Cmd Msg)
stop checkedThingsKeys checkedChannelsIds =
    List.foldr (++)
        []
        (List.map
            (\thingkey ->
                List.map
                    (\channelid ->
                        disconnectWebsocket <|
                            E.object
                                [ ( "channelid", E.string channelid )
                                , ( "thingkey", E.string thingkey )
                                ]
                    )
                    checkedChannelsIds
            )
            checkedThingsKeys
        )
