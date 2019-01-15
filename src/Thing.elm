module Thing exposing (..)

import Http
import Json.Decode as D
import Json.Encode as E


type alias Thing =
    { type_ : String
    , name : String
    , id : String
    , key : String
    }


thingDecoder : D.Decoder Thing
thingDecoder =
    D.map4 Thing
        (D.field "type" D.string)        
        (D.field "name" D.string)
        (D.field "id" D.string)
        (D.field "key" D.string)


thingListDecoder : D.Decoder (List Thing)
thingListDecoder =
    (D.field "things" (D.list thingDecoder))
        

provision : String -> String -> String -> String -> Http.Expect msg -> Cmd msg            
provision url token type_ name msg =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = url
        , body =
            E.object
                [ ("type", E.string type_)
                , ("name", E.string name)
                ]

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


expectRetrieve : (Result Http.Error (List Thing) -> msg) -> Http.Expect msg
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
          case D.decodeString thingListDecoder body of
            Ok value ->
              Ok value

            Err err ->
              Err (Http.BadBody "Account has no things")


remove : String -> String -> String -> Http.Expect msg -> Cmd msg
remove url id token msg =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = url ++ "/" ++ id
        , body = Http.emptyBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
        }
