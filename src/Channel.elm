module Channel exposing (expectChannel, request)

import Http
import Json.Encode as E


type alias Channel =
    { name : String }


encode : Channel -> E.Value
encode channel =
    E.object [ ( "name", E.string channel.name ) ]


request url token name msg =
    { method = "POST"
    , headers = [ Http.header "Authorization" token ]
    , url = url
    , body =
        encode (Channel name)
            |> Http.jsonBody
    , expect = msg
    , timeout = Nothing
    , tracker = Nothing
    }


expectChannel : (Result Http.Error Int -> msg) -> Http.Expect msg
expectChannel toMsg =
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
