module User exposing (..)

import Html exposing  (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E

import Bootstrap.Grid as Grid
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Utilities.Spacing as Spacing

import Error


urls =
    { users = "http://localhost/users"
    , tokens = "http://localhost/tokens"
    }


type alias Model =
    { email : String
    , password : String
    , response : String
    }


initial : Model
initial =
    { email = ""
    , password = ""
    , response = ""
    }


type Msg
    = SubmitEmail String
    | SubmitPassword String
    | Create
    | Created (Result Http.Error Int)
    | GetToken
    | GotToken (Result Http.Error String)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SubmitEmail email ->
            ( { model | email = email }, Cmd.none )

        SubmitPassword password ->
            ( { model | password = password }, Cmd.none )
        Create ->
            ( model
            , request
                model.email
                model.password
                urls.users
                (expectUser Created)
            )

        Created result ->
            case result of
                Ok statusCode ->
                    ( { model | response = "Ok " ++ String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )

        GetToken ->
            ( model
            , request
                model.email
                model.password
                urls.tokens
                (expectToken GotToken)
            )

        GotToken result ->
            case result of
                Ok token ->
                    ( { model | response = "Ok " ++ token }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )                    
        

view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col []
          [ Form.form []
            [ Form.group []
              [ Form.label [ for "myemail" ] [ text "Email address" ]
              , Input.email [ Input.id "myemail", Input.onInput SubmitEmail ]
              ]
            , Form.group []
                [ Form.label [ for "mypwd" ] [ text "Password" ]
                , Input.password [ Input.id "mypwd", Input.onInput SubmitPassword ]
                ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick Create ] [ text "Register" ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick GetToken ] [ text "Token" ]
            ]
          , Html.hr [] []
          , text ("response: " ++ model.response) ]
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


request : String -> String -> String -> Http.Expect msg -> Cmd msg            
request email password url msg =
    Http.request
        { method = "POST"
        , headers = []
        , url = url
        , body =
            encode (User email password)
        |> Http.jsonBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
    }


expectUser : (Result Http.Error Int -> msg) -> Http.Expect msg
expectUser toMsg =
    Http.expectStringResponse toMsg <|
        \response ->
            case response of
                Http.BadUrl_ url ->
                    Err (Http.BadUrl url)

                Http.Timeout_ ->
                    Err Http.Timeout

                Http.NetworkError_ ->
                    Err Http.NetworkError

                Http.BadStatus_ metadata body ->
                    Err (Http.BadStatus metadata.statusCode)

                Http.GoodStatus_ metadata _ ->
                    Ok metadata.statusCode


expectToken : (Result Http.Error String -> msg) -> Http.Expect msg
expectToken toMsg =
    Http.expectStringResponse toMsg <|
        \response ->
            case response of
                Http.BadUrl_ url ->
                    Err (Http.BadUrl url)

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
