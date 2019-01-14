module Channel exposing (..)

import Http
import Json.Encode as E
import Json.Decode as D

type alias Channel =
    { name : String
    , id : String
    }


channelDecoder : D.Decoder Channel
channelDecoder =
    D.map2 Channel
        (D.field "name" D.string)
        (D.field "id" D.string)
    

channelListDecoder : D.Decoder (List Channel)
channelListDecoder =
    (D.field "channels" (D.list channelDecoder))


provision : String -> String -> String -> Http.Expect msg -> Cmd msg
provision url token name msg =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = url
        , body =
            E.object [ ( "name", E.string name ) ]
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


retrieve : String -> String -> Http.Expect msg -> Cmd msg
retrieve url token msg =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" token ]
        , url = url
        , body = Http.emptyBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
        }                    


expectRetrieve : (Result Http.Error (List Channel) -> msg) -> Http.Expect msg
expectRetrieve toMsg =
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
          case D.decodeString channelListDecoder body of
            Ok value ->
              Ok value

            Err err ->
              -- Err (Http.BadBody (D.errorToString err))
              Err (Http.BadBody "Account has no channels")


remove : String -> String -> String -> Http.Expect msg -> Cmd msg
remove url id token msg =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = url ++ id
        , body = Http.emptyBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
        }
