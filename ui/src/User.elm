module User exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Utilities.Spacing as Spacing
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import Url.Builder as B


url =
    { base = "http://localhost"
    , usersPath = [ "users" ]
    , tokensPath = [ "tokens" ]
    }


type alias Model =
    { email : String
    , password : String
    , token : String
    , response : String
    }


initial : Model
initial =
    { email = ""
    , password = ""
    , token = ""
    , response = ""
    }


type Msg
    = SubmitEmail String
    | SubmitPassword String
    | Create
    | Created (Result Http.Error Int)
    | GetToken
    | GotToken (Result Http.Error String)
    | LogOut


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SubmitEmail email ->
            ( { model | email = email }, Cmd.none )

        SubmitPassword password ->
            ( { model | password = password }, Cmd.none )

        Create ->
            ( model
            , create
                model.email
                model.password
                (B.crossOrigin url.base url.usersPath [])
            )

        Created result ->
            case result of
                Ok statusCode ->
                    ( { model | response = String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        GetToken ->
            ( model
            , getToken
                model.email
                model.password
                (B.crossOrigin url.base url.tokensPath [])
            )

        GotToken result ->
            case result of
                Ok token ->
                    ( { model | token = token, response = "" }, Cmd.none )

                Err error ->
                    ( { model | token = "", response = Error.handle error }, Cmd.none )

        LogOut ->
            ( { model | email = "", password = "", token = "", response = "" }, Cmd.none )


view : Model -> Html Msg
view model =
    let
        loggedIn : Bool
        loggedIn =
            if String.length model.token > 0 then
                True

            else
                False
    in
    if loggedIn then
        Grid.container [] []

    else
        Grid.container []
            [ Grid.row []
                [ Grid.col []
                    [ Form.form []
                        [ Form.group []
                            [ Form.label [ for "email" ] [ text "Email address" ]
                            , Input.email [ Input.id "email", Input.onInput SubmitEmail ]
                            ]
                        , Form.group []
                            [ Form.label [ for "pwd" ] [ text "Password" ]
                            , Input.password [ Input.id "pwd", Input.onInput SubmitPassword ]
                            ]
                        , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick Create ] [ text "Register" ]
                        , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick GetToken ] [ text "Log in" ]
                        ]
                    ]
                ]
            , Helpers.response model.response
            ]


type alias User =
    { email : String
    , password : String
    }


encode : User -> E.Value
encode user =
    E.object
        [ ( "email", E.string user.email )
        , ( "password", E.string user.password )
        ]


decoder : D.Decoder User
decoder =
    D.map2 User
        (D.field "email" D.string)
        (D.field "password" D.string)


create : String -> String -> String -> Cmd Msg
create email password u =
    Http.request
        { method = "POST"
        , headers = []
        , url = u
        , body =
            encode (User email password)
                |> Http.jsonBody
        , expect = expectUser Created
        , timeout = Nothing
        , tracker = Nothing
        }


expectUser : (Result Http.Error Int -> Msg) -> Http.Expect Msg
expectUser toMsg =
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


getToken : String -> String -> String -> Cmd Msg
getToken email password u =
    Http.request
        { method = "POST"
        , headers = []
        , url = u
        , body =
            encode (User email password)
                |> Http.jsonBody
        , expect = expectToken GotToken
        , timeout = Nothing
        , tracker = Nothing
        }


expectToken : (Result Http.Error String -> Msg) -> Http.Expect Msg
expectToken toMsg =
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

                Http.GoodStatus_ metadata body ->
                    case D.decodeString (D.field "token" D.string) body of
                        Ok value ->
                            Ok value

                        Err err ->
                            Err (Http.BadBody (D.errorToString err))
