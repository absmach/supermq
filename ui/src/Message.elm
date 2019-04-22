-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


port module Message exposing (Model, Msg(..), initial, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Radio as Radio
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Channel
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
    { message : String
    , thingkey : String
    , response : String
    , things : Thing.Model
    , channels : Channel.Model
    , thingid : String
    , checkedChannelsIds : List String
    , websocketData : List String
    }


initial : Model
initial =
    { message = ""
    , thingkey = ""
    , response = ""
    , things = Thing.initial
    , channels = Channel.initial
    , thingid = ""
    , checkedChannelsIds = []
    , websocketData = []
    }


type Msg
    = SubmitMessage String
    | SendMessage
    | WebsocketSend
    | WebsocketMsg String
    | SentMessage (Result Http.Error String)
    | ThingMsg Thing.Msg
    | ChannelMsg Channel.Msg
    | SelectedThing String String Channel.Msg
    | CheckChannel String


resetSent : Model -> Model
resetSent model =
    { model | message = "", thingkey = "", response = "", thingid = "" }


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitMessage message ->
            ( { model | message = message }, Cmd.none )

        SendMessage ->
            ( model
            , Cmd.batch
                (List.map
                    (\channelid -> send channelid model.thingkey model.message)
                    model.checkedChannelsIds
                )
            )

        WebsocketSend ->
            ( model
            , Cmd.batch
                (List.map
                    (\channelid ->
                        websocketOut <|
                            E.object
                                [ ( "channelid", E.string channelid )
                                , ( "thingkey", E.string model.thingkey )
                                , ( "message", E.string model.message )
                                ]
                    )
                    model.checkedChannelsIds
                )
            )

        SentMessage result ->
            case result of
                Ok statusCode ->
                    ( { model | response = statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        WebsocketMsg data ->
            ( { model | websocketData = data :: Helpers.resetList model.websocketData 5 }, Cmd.none )

        ThingMsg subMsg ->
            updateThing model subMsg token

        ChannelMsg subMsg ->
            updateChannel model subMsg token

        SelectedThing thingid thingkey channelMsg ->
            updateChannel { model | thingid = thingid, thingkey = thingkey, checkedChannelsIds = [] } (Channel.RetrieveChannelsForThing thingid) token

        CheckChannel id ->
            ( { model | checkedChannelsIds = Helpers.checkEntity id model.checkedChannelsIds }, Cmd.none )


updateThing : Model -> Thing.Msg -> String -> ( Model, Cmd Msg )
updateThing model msg token =
    let
        ( updatedThing, thingCmd ) =
            Thing.update msg model.things token
    in
    ( { model | things = updatedThing }, Cmd.map ThingMsg thingCmd )


updateChannel : Model -> Channel.Msg -> String -> ( Model, Cmd Msg )
updateChannel model msg token =
    let
        ( updatedChannel, channelCmd ) =
            Channel.update msg model.channels token
    in
    ( { model | channels = updatedChannel }, Cmd.map ChannelMsg channelCmd )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ websocketIn WebsocketMsg
        , websocketState WebsocketMsg
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                (Helpers.appendIf (model.things.things.total > model.things.limit)
                    [ Helpers.genCardConfig faIcons.things "Things" (genThingRows model.things.things.list) ]
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
                [ Card.config []
                    |> Card.headerH3 [] [ div [ class "table_header" ] [ i [ style "margin-right" "15px", class faIcons.messages ] [], text "Message" ] ]
                    |> Card.block []
                        [ Block.custom
                            (Form.form []
                                [ Form.group []
                                    [ Input.text [ Input.id "message", Input.onInput SubmitMessage ]
                                    ]
                                , Button.button [ Button.secondary, Button.attrs [ Spacing.ml1 ], Button.onClick SendMessage ] [ text "Send" ]
                                , Button.button [ Button.secondary, Button.attrs [ Spacing.ml1 ], Button.onClick WebsocketSend ] [ text "Websocket" ]
                                ]
                            )
                        ]
                    |> Card.view
                ]
            ]
        , Helpers.response model.response
        , Helpers.genOrderedList model.websocketData
        ]


genThingRows : List Thing.Thing -> List (Table.Row Msg)
genThingRows things =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ label [] [ text (Helpers.parseString thing.name) ] ]
                , Table.td [] [ text thing.id ]
                , Table.td [] [ input [ type_ "radio", onClick (SelectedThing thing.id thing.key (Channel.RetrieveChannelsForThing thing.id)), name "things" ] [] ]
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


send : String -> String -> String -> Cmd Msg
send channelid thingkey message =
    HttpMF.request
        (B.relative [ "http", paths.channels, channelid, paths.messages ] [])
        "POST"
        thingkey
        (Http.stringBody "application/json" message)
        SentMessage
