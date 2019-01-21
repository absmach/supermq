module Message exposing (Model, Msg(..), expectMessage, initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Utilities.Spacing as Spacing
import Error
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Url.Builder as B


url =
    { base = "http://localhost"
    }


type alias Model =
    { message : String
    , token : String
    , channel : String
    , response : String
    }


initial : Model
initial =
    { message = ""
    , token = ""
    , channel = ""
    , response = ""
    }


type Msg
    = SubmitMessage String
    | SubmitToken String
    | SubmitChannel String
    | SendMessage
    | SentMessage (Result Http.Error Int)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SubmitMessage message ->
            ( { model | message = message }, Cmd.none )

        SubmitChannel channel ->
            ( { model | channel = channel }, Cmd.none )

        SubmitToken token ->
            ( { model | token = token }, Cmd.none )

        SendMessage ->
            ( model
            , Http.request
                { method = "POST"
                , headers = [ Http.header "Authorization" model.token ]
                , url = B.crossOrigin url.base [ "http", "channels", model.channel, "messages" ] []
                , body = Http.stringBody "application/json" model.message
                , expect = expectMessage SentMessage
                , timeout = Nothing
                , tracker = Nothing
                }
            )

        SentMessage result ->
            case result of
                Ok statusCode ->
                    ( { model | response = "Ok " ++ String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )


view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col []
            [ Form.form []
                [ Form.group []
                    [ Form.label [ for "mychan" ] [ text "Channel" ]
                    , Input.email [ Input.id "mychan", Input.onInput SubmitChannel ]
                    ]
                , Form.group []
                    [ Form.label [ for "mytoken" ] [ text "Token" ]
                    , Input.text [ Input.id "mytoken", Input.onInput SubmitToken ]
                    ]
                , Form.group []
                    [ Form.label [ for "mymessage" ] [ text "Message" ]
                    , Input.text [ Input.id "mymessage", Input.onInput SubmitMessage ]
                    ]
                , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick SendMessage ] [ text "Send" ]
                ]
            , Html.hr [] []
            , text ("response: " ++ model.response)
            ]
        ]


expectMessage : (Result Http.Error Int -> msg) -> Http.Expect msg
expectMessage toMsg =
    Http.expectStringResponse toMsg <|
        \response ->
            case response of
                Http.BadUrl_ u ->
                    Err (Http.BadUrl u)

                Http.Timeout_ ->
                    Err Http.Timeout

                Http.NetworkError_ ->
                    Err Http.NetworkError

                Http.BadStatus_ metadata body ->
                    Err (Http.BadStatus metadata.statusCode)

                Http.GoodStatus_ metadata _ ->
                    Ok metadata.statusCode
