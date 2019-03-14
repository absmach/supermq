-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


module Main exposing (Model, Msg(..), init, main, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.CDN as CDN
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Fieldset as Fieldset
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Radio as Radio
import Bootstrap.Form.Select as Select
import Bootstrap.Form.Textarea as Textarea
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Text as Text
import Bootstrap.Utilities.Spacing as Spacing
import Browser
import Browser.Navigation as Nav
import Channel
import Connection
import Debug exposing (log)
import Error
import Helpers exposing (Globals, fontAwesome)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode exposing (Decoder, field, string)
import Json.Encode as Encode
import Message
import Thing
import Url
import Url.Parser as UrlParser exposing ((</>))
import User
import Version



-- MAIN


type alias Flags =
    { protocol : String, host : String, port_ : String }


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type alias Model =
    { key : Nav.Key
    , user : User.Model
    , version : Version.Model
    , channel : Channel.Model
    , thing : Thing.Model
    , connection : Connection.Model
    , message : Message.Model
    , view : String
    , globals : Globals
    }


init : Flags -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        baseURL =
            flags.protocol ++ "://" ++ flags.host ++ ":" ++ flags.port_
    in
    ( Model key
        User.initial
        Version.initial
        Channel.initial
        Thing.initial
        Connection.initial
        Message.initial
        (parse url)
        (Globals baseURL "")
    , Cmd.none
    )



-- URL PARSER


type alias Route =
    ( String, Maybe String )


parse : Url.Url -> String
parse url =
    UrlParser.parse
        (UrlParser.map Tuple.pair (UrlParser.string </> UrlParser.fragment identity))
        url
        |> (\route ->
                case route of
                    Just r ->
                        Tuple.first r

                    Nothing ->
                        ""
           )


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | UserMsg User.Msg
    | VersionMsg Version.Msg
    | ChannelMsg Channel.Msg
    | ThingMsg Thing.Msg
    | ConnectionMsg Connection.Msg
    | MessageMsg Message.Msg
    | Dashboard
    | Channels
    | Things
    | Connection
    | Messages



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Cmd.none )

        UrlChanged url ->
            ( { model | view = parse url }
            , Cmd.none
            )

        UserMsg subMsg ->
            updateUser model subMsg

        VersionMsg subMsg ->
            updateVersion model subMsg

        ChannelMsg subMsg ->
            updateChannel model subMsg

        ThingMsg subMsg ->
            updateThing model subMsg

        ConnectionMsg subMsg ->
            updateConnection model subMsg

        MessageMsg subMsg ->
            updateMessage model subMsg

        Dashboard ->
            ( { model | view = "dashboard" }, Cmd.none )

        Things ->
            ( { model | view = "things" }, Cmd.none )

        Channels ->
            ( { model | view = "channels" }, Cmd.none )

        Connection ->
            ( { model | view = "connection" }
            , Cmd.batch
                [ Tuple.second (updateConnection model (Connection.ThingMsg Thing.RetrieveThings))
                , Tuple.second (updateConnection model (Connection.ChannelMsg Channel.RetrieveChannels))
                ]
            )

        Messages ->
            updateMessage { model | view = "messages" } (Message.ThingMsg Thing.RetrieveThings)


logIn : Model -> Version.Msg -> Thing.Msg -> Channel.Msg -> ( Model, Cmd Msg )
logIn model versionMsg thingMsg channelMsg =
    ( model
    , Cmd.batch
        [ Tuple.second (updateVersion model versionMsg)
        , Tuple.second (updateThing model thingMsg)
        , Tuple.second (updateChannel model channelMsg)
        ]
    )


loggedIn : Model -> Bool
loggedIn model =
    String.length model.globals.token > 0


updateUser : Model -> User.Msg -> ( Model, Cmd Msg )
updateUser model msg =
    let
        ( updatedUser, userCmd ) =
            User.update model.globals msg model.user

        globs =
            model.globals
    in
    if String.length updatedUser.token > 0 then
        if not (loggedIn model) then
            logIn { model | user = updatedUser, view = "dashboard", globals = { globs | token = updatedUser.token } } Version.GetVersion Thing.RetrieveThings Channel.RetrieveChannels

        else
            ( { model | user = updatedUser, globals = { globs | token = updatedUser.token } }, Cmd.none )

    else if loggedIn model then
        ( { model | user = User.initial, globals = { globs | token = "" } }, Cmd.map UserMsg userCmd )

    else
        ( { model | user = updatedUser }, Cmd.map UserMsg userCmd )


updateVersion : Model -> Version.Msg -> ( Model, Cmd Msg )
updateVersion model msg =
    let
        ( updatedVersion, versionCmd ) =
            Version.update model.globals msg model.version
    in
    ( { model | version = updatedVersion }, Cmd.map VersionMsg versionCmd )


updateThing : Model -> Thing.Msg -> ( Model, Cmd Msg )
updateThing model msg =
    let
        ( updatedThing, thingCmd ) =
            Thing.update model.globals msg model.thing
    in
    ( { model | thing = updatedThing }, Cmd.map ThingMsg thingCmd )


updateChannel : Model -> Channel.Msg -> ( Model, Cmd Msg )
updateChannel model msg =
    let
        ( updatedChannel, channelCmd ) =
            Channel.update model.globals msg model.channel
    in
    ( { model | channel = updatedChannel }, Cmd.map ChannelMsg channelCmd )


updateConnection : Model -> Connection.Msg -> ( Model, Cmd Msg )
updateConnection model msg =
    let
        ( updatedConnection, connectionCmd ) =
            Connection.update model.globals msg model.connection
    in
    ( { model | connection = updatedConnection }, Cmd.map ConnectionMsg connectionCmd )


updateMessage : Model -> Message.Msg -> ( Model, Cmd Msg )
updateMessage model msg =
    let
        ( updatedMessage, messageCmd ) =
            Message.update model.globals msg model.message
    in
    ( { model | message = updatedMessage }, Cmd.map MessageMsg messageCmd )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Sub.map ThingMsg (Thing.subscriptions model.thing)
        , Sub.map UserMsg (User.subscriptions model.user)
        ]



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateflux"
    , body =
        let
            buttonAttrs =
                Button.attrs [ style "text-align" "left" ]

            menu =
                if loggedIn model then
                    [ ButtonGroup.linkButton [ Button.primary, Button.onClick Dashboard, buttonAttrs ] [ i [ class "fas fa-chart-bar" ] [], text " Dashboard" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Things, buttonAttrs ] [ i [ class "fas fa-sitemap" ] [], text " Things" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Channels, buttonAttrs ] [ i [ class "fas fa-broadcast-tower" ] [], text " Channels" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Connection, buttonAttrs ] [ i [ class "fas fa-plug" ] [], text " Connection" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Messages, buttonAttrs ] [ i [ class "far fa-paper-plane" ] [], text " Messages" ]
                    ]

                else
                    []

            header =
                Html.map UserMsg (User.view (loggedIn model) model.user)

            content =
                if loggedIn model then
                    case model.view of
                        "dashboard" ->
                            dashboard model

                        "channels" ->
                            Html.map ChannelMsg (Channel.view model.channel)

                        "things" ->
                            Html.map ThingMsg (Thing.view model.thing)

                        "connection" ->
                            Html.map ConnectionMsg (Connection.view model.connection)

                        "messages" ->
                            Html.map MessageMsg (Message.view model.message)

                        _ ->
                            dashboard model

                else
                    Grid.container [] []
        in
        [ Grid.containerFluid []
            [ fontAwesome
            , Grid.row [ Row.attrs [ style "height" "100vh" ] ]
                [ Grid.col
                    [ Col.attrs
                        [ style "background-color" "#113f67"
                        , style "padding" "0"
                        , style "color" "white"
                        ]
                    ]
                    [ Grid.row []
                        [ Grid.col
                            [ Col.attrs [] ]
                            [ h3 [ class "title" ] [ text "MAINFLUX" ] ]
                        ]
                    , Grid.row []
                        [ Grid.col
                            [ Col.attrs [] ]
                            [ ButtonGroup.linkButtonGroup
                                [ ButtonGroup.vertical
                                , ButtonGroup.attrs [ style "width" "100%" ]
                                ]
                                menu
                            ]
                        ]
                    ]
                , Grid.col
                    [ Col.xs10
                    , Col.attrs []
                    ]
                    [ header
                    , Grid.row []
                        [ Grid.col
                            [ Col.attrs [] ]
                            [ content ]
                        ]
                    ]
                ]
            ]
        ]
    }


dashboard : Model -> Html Msg
dashboard model =
    Grid.container
        []
        [ Grid.row []
            [ Grid.col []
                [ Card.deck (cardList model)
                ]
            ]
        ]


cardList : Model -> List (Card.Config Msg)
cardList model =
    [ Card.config
        [ Card.secondary
        , Card.textColor Text.white
        ]
        |> Card.headerH3 [] [ text "Version" ]
        |> Card.block []
            [ Block.titleH4 [] [ text model.version.version ] ]
    , Card.config
        [ Card.info
        , Card.textColor Text.white
        ]
        |> Card.headerH3 [] [ text "Things" ]
        |> Card.block []
            [ Block.titleH4 [] [ text (String.fromInt model.thing.things.total) ]
            , Block.custom <|
                Button.button [ Button.light, Button.onClick Things ] [ text "Manage things" ]
            ]
    , Card.config []
        |> Card.headerH3 [] [ text "Channels" ]
        |> Card.block []
            [ Block.titleH4 [] [ text (String.fromInt model.channel.channels.total) ]
            , Block.custom <|
                Button.button [ Button.dark, Button.onClick Channels ] [ text "Manage channels" ]
            ]
    ]
