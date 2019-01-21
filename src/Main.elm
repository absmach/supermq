module Main exposing (Model, Msg(..), init, main, menuButtons, subscriptions, update, view)

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
    , route : Maybe Route
    , version : Version.Model
    , user : User.Model
    , channel : Channel.Model
    , thing : Thing.Model
    , connection : Connection.Model
    , message : Message.Model
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( Model key
        (UrlParser.parse parser url)
        Version.initial
        User.initial
        Channel.initial
        Thing.initial
        Connection.initial
        Message.initial
    , Cmd.none
    )



-- URL PARSER


type alias Route =
    ( String, Maybe String )


parser : UrlParser.Parser (Route -> a) a
parser =
    UrlParser.map Tuple.pair (UrlParser.string </> UrlParser.fragment identity)



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | VersionMsg Version.Msg
    | UserMsg User.Msg
    | ChannelMsg Channel.Msg
    | ThingMsg Thing.Msg
    | ConnectionMsg Connection.Msg
    | MessageMsg Message.Msg


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
            ( { model | route = UrlParser.parse parser url }
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
            content =
                case model.route of
                    Just route ->
                        if (loggedIn model) then
                            case Tuple.first route of
                                "version" ->
                                    Html.map VersionMsg (Version.view model.version)

                                "account" ->
                                    Html.map UserMsg (User.view model.user)

                                "channel" ->
                                    Html.map ChannelMsg (Channel.view model.channel)

                                "things" ->
                                    Html.map ThingMsg (Thing.view model.thing)

                                "connection" ->
                                    Html.map ConnectionMsg (Connection.view model.connection)

                                "messages" ->
                                    Html.map MessageMsg (Message.view model.message)

                                _ ->
                                    h3 [] [ text "Welcome to Gateflux" ]
                        else
                            Html.map UserMsg (User.view model.user)
                                
                    Nothing ->
                        Html.map UserMsg (User.view model.user)
        in
        [ -- we use Bootstrap container defined at http://elm-bootstrap.info/grid
          Grid.container []
            [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS
            , Grid.row []
                [ Grid.col [] [ h1 [] [ a [href "/account"] [text "Gateflux"] ] ] ]
            , Grid.row []
                [ Grid.col []
                    [ -- In this column we put the button group defined below
                      ButtonGroup.linkButtonGroup [ ButtonGroup.vertical ] (menuButtons model)
                    ]
                , Grid.col [ Col.xs10 ]
                    [ content
                    ]
                ]
            ]
        ]
    }


menuButtons : Model -> List (ButtonGroup.LinkButtonItem msg)
menuButtons model =
    if (loggedIn model) then
        [ ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/account" ] ] [ text "Account" ]
        , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/channel" ] ] [ text "Channels" ]
        , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/things" ] ] [ text "Things" ]
        , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/connection" ] ] [ text "Connection" ]
        , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/messages" ] ] [ text "Messages" ]
        , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/version" ] ] [ text "Version" ]        
        ]
    else
        [ ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/account" ] ] [ text "Account" ]
        ]        


loggedIn : Model -> Bool
loggedIn model =
    if String.length model.user.token > 0 then
        True
    else
        False
