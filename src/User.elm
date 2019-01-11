module User exposing (User, decoder, encode, expectToken, expectUser, request)

import Http
import Json.Decode as D
import Json.Encode as E


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
