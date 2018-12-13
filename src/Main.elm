module Main exposing (main)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.CDN as CDN
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Browser
import Browser.Navigation as Nav
import Debug exposing (log)
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode exposing (Decoder, field, string)
import Json.Encode as Encode
import Url
import User



-- MAIN


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type alias Model =
    { key : Nav.Key
    , url : Url.Url
    , res : String
    , email : String
    , password : String
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    ( Model key url "" "" "", Cmd.none )



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | GotJSON (Result Http.Error String)
    | SubmitEmail String
    | SubmitPassword String
    | SubmitUser
    | GotUser (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model
                    , Http.get
                        { url = href
                        , expect = Http.expectJson GotJSON jsonDecoder
                        }
                    )

        UrlChanged url ->
            ( { model | url = url }
            , Cmd.none
            )

        GotJSON result ->
            case result of
                Ok text ->
                    ( { model | res = text }, Cmd.none )

                Err _ ->
                    log "err"
                        ( model, Cmd.none )

        SubmitEmail email ->
            ( { model | email = email }, Cmd.none )

        SubmitPassword password ->
            ( { model | password = password }, Cmd.none )

        SubmitUser ->
            ( model, requestUser model )

        -- ( model, Cmd.none )
        GotUser _ ->
            ( model, Cmd.none )


jsonDecoder : Decoder String
jsonDecoder =
    field "version" string



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "URL Interceptor"
    , body =
        [ Grid.container []
            [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS,
            , Button.linkButton
                [ Button.primary, Button.attrs [ href "http://localhost/version" ] ]
                [ text "Version" ]
            , Button.linkButton
                [ Button.primary, Button.attrs [ href "#" ] ]
                [ text "User" ]
            , Grid.row []
                [ Grid.col []
                    [ text ("version: " ++ model.res) ]
                , Grid.col []
                    [ text ("email: " ++ model.email) ]
                , Grid.col []
                    [ text ("password: " ++ model.password) ]
                ]
            , Form.form []
                [ Form.group []
                    [ Form.label [ for "myemail" ] [ text "Email address" ]
                    , Input.email [ Input.id "myemail", Input.onInput SubmitEmail ]
                    , Form.help [] [ text "We'll never share your email with anyone else." ]
                    ]
                , Form.group []
                    [ Form.label [ for "mypwd" ] [ text "Password" ]
                    , Input.password [ Input.id "mypwd", Input.onInput SubmitPassword ]
                    ]
                , Button.button [ Button.primary, Button.onClick SubmitUser ] [ text "Submit" ]
                ]
            ]
        ]
    }



-- HELPERS


requestUser : Model -> Cmd Msg
requestUser model =
    Http.request
        { method = "POST"
        , headers = []

        -- , headers =
        --     [ Http.header "Origin" "http://localhost:8000/"
        --     , Http.header "Access-Control-Request-Method" "POST"
        --     , Http.header "Access-Control-Request-Headers" "X-Custom-Header"
        --     ]
        , url = "http://localhost/users"

        -- , body = jsonBody (User.user (User.User model.email model.password))
        , body = Http.jsonBody (User.user (User.User model.email model.password))
        , expect = Http.expectWhatever GotUser
        , timeout = Nothing
        , tracker = Nothing
        }


jsonBody : Encode.Value -> Http.Body
jsonBody value =
    Http.stringBody "application/json" (Encode.encode 0 value)
