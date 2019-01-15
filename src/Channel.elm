module Channel exposing (..)

import Http
import Html exposing  (..)
import Html.Attributes exposing (..)
import Json.Encode as E
import Json.Decode as D

import Bootstrap.Grid as Grid
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Utilities.Spacing as Spacing

import Error


urls =
    {
        channels = "http://localhost/channels"
    }
   

type alias Model =
    { channel : String
    , token : String
    , response : String
    }


initial : Model
initial =
    { channel = ""
    , token = ""
    , response = ""
    }


type Msg
    = SubmitChannel String
    | SubmitToken String
    | ProvisionChannel
    | ProvisionedChannel (Result Http.Error Int)
    | RetrieveChannel
    | RetrievedChannel (Result Http.Error (List Channel))
    | RemoveChannel


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SubmitChannel channel ->
            ( { model | channel = channel }, Cmd.none )

        SubmitToken token ->
            ( { model | token = token }, Cmd.none )

        ProvisionChannel ->
            ( model
            , provision
                urls.channels
                model.token
                model.channel
                (expectProvision ProvisionedChannel)
            )

        ProvisionedChannel result ->
            case result of
                Ok statusCode ->
                    ( { model | response = "Ok " ++ String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )

        RetrieveChannel ->
            ( model
            , retrieve
                urls.channels
                model.token
                (expectRetrieve RetrievedChannel)
            )

        RetrievedChannel result ->
            case result of
                Ok channels ->
                    ( { model | response = channelsToString channels }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )
            
        RemoveChannel ->
            ( model
            , remove
                urls.channels
                model.channel                     
                model.token
                (expectProvision ProvisionedChannel)                     
            )            


view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col []
          [ Form.form []
            [ Form.group []
              [ Form.label [ for "mychan" ] [ text "Name (Provision) or id (Remove)" ]
              , Input.email [ Input.id "mychan", Input.onInput SubmitChannel ]
              ]
            , Form.group []
                [ Form.label [ for "mytoken" ] [ text "Token" ]
                , Input.text [ Input.id "mytoken", Input.onInput SubmitToken ]
                ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick ProvisionChannel ] [ text "Provision" ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RetrieveChannel ] [ text "Retrieve" ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RemoveChannel ] [ text "Remove" ]
            ]
          , Html.hr [] []
          , text ("response: " ++ model.response)
          ]
        ]


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
        , url = url ++ "/" ++ id
        , body = Http.emptyBody
        , expect = msg
        , timeout = Nothing
        , tracker = Nothing
        }

-- HELPERS

                
channelsToString : List Channel -> String
channelsToString channels =
    List.map
        (\channel -> channel.name ++ " " ++ channel.id ++ "; ")
        channels
        |> String.concat
        
