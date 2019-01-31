module Main exposing (Model, Msg(..), init, main, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.CDN as CDN
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
import Bootstrap.Utilities.Spacing as Spacing
import Browser
import Browser.Navigation as Nav
import Channel
import Connection
import Error
import Html exposing (..)
import Html.Attributes exposing (..)
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


main : Program () Model Msg
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
    , version : Version.Model
    , user : User.Model
    , channel : Channel.Model
    , thing : Thing.Model
    , connection : Connection.Model
    , message : Message.Model
    , view : String
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( Model key
        Version.initial
        User.initial
        Channel.initial
        Thing.initial
        Connection.initial
        Message.initial
        (parse url)
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
                        "account"
           )


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | VersionMsg Version.Msg
    | UserMsg User.Msg
    | ChannelMsg Channel.Msg
    | ThingMsg Thing.Msg
    | ConnectionMsg Connection.Msg
    | MessageMsg Message.Msg
    | Account
    | Channels
    | Things
    | Connection
    | Messages
    | Version


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

        VersionMsg subMsg ->
            let
                ( updatedVersion, versionCmd ) =
                    Version.update subMsg model.version
            in
            ( { model | version = updatedVersion }, Cmd.map VersionMsg versionCmd )

        UserMsg subMsg ->
            let
                ( updatedUser, userCmd ) =
                    User.update subMsg model.user
            in
            ( { model | user = updatedUser }, Cmd.map UserMsg userCmd )

        ChannelMsg subMsg ->
            let
                ( updatedChannel, channelCmd ) =
                    Channel.update subMsg model.channel model.user.token
            in
            ( { model | channel = updatedChannel }, Cmd.map ChannelMsg channelCmd )

        ThingMsg subMsg ->
            let
                ( updatedThing, thingCmd ) =
                    Thing.update subMsg model.thing model.user.token
            in
            ( { model | thing = updatedThing }, Cmd.map ThingMsg thingCmd )

        ConnectionMsg subMsg ->
            let
                ( updatedConnection, connectionCmd ) =
                    Connection.update subMsg model.connection model.user.token
            in
            ( { model | connection = updatedConnection }, Cmd.map ConnectionMsg connectionCmd )

        MessageMsg subMsg ->
            let
                ( updatedMessage, messageCmd ) =
                    Message.update subMsg model.message
            in
            ( { model | message = updatedMessage }, Cmd.map MessageMsg messageCmd )

        Account ->
            ( { model | view = "account" }, Cmd.none )

        Channels ->
            let
                ( updatedChannel, channelCmd ) =
                    Channel.update Channel.RetrieveChannels model.channel model.user.token
            in
            ( { model | view = "channels" }, Cmd.map ChannelMsg channelCmd )

        Things ->
            let
                ( updatedThing, thingCmd ) =
                    Thing.update Thing.RetrieveThings model.thing model.user.token
            in
            ( { model | view = "things" }, Cmd.map ThingMsg thingCmd )

        Connection ->
            ( { model | view = "connection" }, Cmd.none )

        Messages ->
            ( { model | view = "messages" }, Cmd.none )

        Version ->
            let
                ( updatedVersion, versionCmd ) =
                    Version.update Version.GetVersion model.version
            in
            ( { model | view = "version", version = updatedVersion }, Cmd.map VersionMsg versionCmd )



-- Menu subMsg ->
--     ( model, Cmd.none )
-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateflux"
    , body =
        let
            loggedIn : Bool
            loggedIn =
                if String.length model.user.token > 0 then
                    True

                else
                    False

            buttonAttrs =
                Button.attrs [ style "color" "white" ]

            menu =
                if loggedIn then
                    [ ButtonGroup.linkButton [ Button.primary, Button.onClick Account, buttonAttrs ] [ text "Account" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Channels, buttonAttrs ] [ text "Channels" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Things, buttonAttrs ] [ text "Things" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Connection, buttonAttrs ] [ text "Connection" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Messages, buttonAttrs ] [ text "Messages" ]
                    , ButtonGroup.linkButton [ Button.primary, Button.onClick Version, buttonAttrs ] [ text "Version" ]
                    ]

                else
                    []

            content =
                if loggedIn then
                    case model.view of
                        "version" ->
                            Html.map VersionMsg (Version.view model.version)

                        "account" ->
                            Html.map UserMsg (User.view model.user)

                        "channels" ->
                            Html.map ChannelMsg (Channel.view model.channel)

                        "things" ->
                            Html.map ThingMsg (Thing.view model.thing)

                        "connection" ->
                            Html.map ConnectionMsg (Connection.view model.connection)

                        "messages" ->
                            Html.map MessageMsg (Message.view model.message)

                        _ ->
                            Html.map UserMsg (User.view model.user)

                else
                    Html.map UserMsg (User.view model.user)
        in
        [ -- we use Bootstrap container defined at http://elm-bootstrap.info/grid
          Grid.container []
            [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS
            , Grid.row []
                [ Grid.col [] [ h1 [] [ text "Gateflux" ] ] ]
            , Grid.row []
                [ Grid.col []
                    [ -- In this column we put the button group defined below
                      ButtonGroup.linkButtonGroup [ ButtonGroup.vertical ] menu
                    ]
                , Grid.col [ Col.xs10 ]
                    [ content
                    ]
                ]
            ]
        ]
    }
