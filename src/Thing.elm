module Thing exposing (..)

import Http
import Json.Decode as D
import Json.Encode as E


type alias Thing =
    { type_ : String
    , name : String
    }


encode : Thing -> E.Value
encode thing =
    E.object
        [ ("type", E.string thing.type_)
        , ("name", E.string thing.name)
        ]


provision : String -> String -> String -> String -> Http.Expect msg -> Cmd msg            
provision url token type_ name msg =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = url
        , body =
            encode (Thing type_ name)
        |> Http.jsonBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
    }


expectProvision : (Result Http.Error Int -> msg) -> Http.Expect msg
expectProvision toMsg =
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
