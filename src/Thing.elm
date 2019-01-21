module Thing exposing (..)

import Http
import Html exposing  (..)
import Html.Attributes exposing (..)
import Json.Encode as E
import Json.Decode as D
import Url.Builder as B

import Bootstrap.Grid as Grid
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Utilities.Spacing as Spacing

import Error


path =
    { offset = "0"
    , limit = "10"
    }


url =
    { base = "http://localhost"        
    , path = [ "things" ]
    }
   

type alias Model =
    { name : String
    , type_ : String
    , offset : String
    , limit : String              
    , response : String
    }


initial : Model
initial =
    { name = ""
    , type_ = ""
    , offset = path.offset
    , limit = path.limit              
    , response = ""
    }


type Msg
    = SubmitType String
    | SubmitName String
    | SubmitOffset String
    | SubmitLimit String      
    | ProvisionThing
    | ProvisionedThing (Result Http.Error Int)      
    | RetrieveThing
    | RetrievedThing (Result Http.Error (List Thing))
    | RemoveThing



update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitType type_ ->
            ( { model | type_ = type_ }, Cmd.none )

        SubmitName name ->
            ( { model | name = name }, Cmd.none )

        SubmitOffset offset ->
            ( { model | offset = offset }, Cmd.none )

        SubmitLimit limit ->
            ( { model | limit = limit }, Cmd.none )

        ProvisionThing ->
            ( model
            , provision
                (B.crossOrigin url.base url.path [])
                token
                model.type_
                model.name
            )

        ProvisionedThing result ->
            case result of
                Ok statusCode ->
                    ( { model | response = "Ok " ++ String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )
            
        RetrieveThing ->
            ( model
            , retrieve
                (B.crossOrigin url.base url.path (buildQueryParamList model))
                token
            )
            

        RetrievedThing result ->
            case result of
                Ok things ->
                    ( { model | response = thingsToString things }, Cmd.none )
                    
                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )
            
        RemoveThing ->
            ( model
            , remove
                (B.crossOrigin url.base (List.append url.path [ model.name ]) [])
                token
            )            



view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col []
          [ Form.form []
            [ Form.group []
              [ Form.label [ for "type" ] [ text "Type" ]
              , Input.text [ Input.id "type", Input.onInput SubmitType ]
              ]
            , Form.group []
                [ Form.label [ for "name" ] [ text "Name (Provision) or id (Remove)" ]
                , Input.text [ Input.id "name", Input.onInput SubmitName ]
                ]
            , Form.group []
                [ Form.label [ for "offset" ] [ text "Offset" ]
                , Input.text [ Input.id "offset", Input.onInput SubmitOffset ]
                ]
            , Form.group []
                [ Form.label [ for "limit" ] [ text "Limit" ]
                , Input.text [ Input.id "limit", Input.onInput SubmitLimit ]
                ]                                
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick ProvisionThing ] [ text "Provision" ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RetrieveThing ] [ text "Retrieve" ]
            , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RemoveThing ] [ text "Remove" ]
            ]
          , Html.hr [] []
          , text ("response: " ++ model.response) ]
        ]

        
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
        

provision : String -> String -> String -> String -> Cmd Msg            
provision u token type_ name =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body =
            E.object
                [ ("type", E.string type_)
                , ("name", E.string name)
                ]

        |> Http.jsonBody
        , expect = expectProvision ProvisionedThing
        , timeout = Nothing
        , tracker = Nothing
    }


expectProvision : (Result Http.Error Int -> Msg) -> Http.Expect Msg
expectProvision toMsg =
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


retrieve : String -> String -> Cmd Msg
retrieve u token =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = expectRetrieve RetrievedThing
        , timeout = Nothing
        , tracker = Nothing
        }                    


expectRetrieve : (Result Http.Error (List Thing) -> Msg) -> Http.Expect Msg
expectRetrieve toMsg =
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
          case D.decodeString thingListDecoder body of
            Ok value ->
              Ok value

            Err err ->
              Err (Http.BadBody "Account has no things")


remove : String -> String -> Cmd Msg
remove u token =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = expectProvision ProvisionedThing
        , timeout = Nothing
        , tracker = Nothing
        }



-- HELPERS


thingsToString : List Thing -> String
thingsToString things =
    List.map
        (\thing -> thing.id ++ " " ++ thing.type_ ++ " " ++ thing.name ++ " " ++ thing.key ++ "; ")
        things
        |> String.concat


buildQueryParamList : Model -> List B.QueryParameter
buildQueryParamList model =
    List.map 
        (\tpl ->
             case (String.toInt (Tuple.second tpl)) of
                 Just n ->
                     B.int (Tuple.first tpl) n
                            
                 Nothing ->
                     if (Tuple.first tpl) == "offset" then
                         B.string (Tuple.first tpl) path.offset
                     else
                         B.string (Tuple.first tpl) path.limit)
        [("offset", model.offset), ("limit", model.limit)]           
