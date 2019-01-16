module Thing exposing (..)

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
        things = "http://localhost/things"
    }
   

type alias Model =
    { name : String
    , type_ : String
    , token : String
    , response : String
    }


initial : Model
initial =
    { name = ""
    , type_ = ""
    , token = ""
    , response = ""
    }


type Msg
    = SubmitToken String
    | SubmitType String
    | SubmitName String
    | ProvisionThing
    | ProvisionedThing (Result Http.Error Int)      
    | RetrieveThing
    | RetrievedThing (Result Http.Error (List Thing))
    | RemoveThing



update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SubmitToken token ->
            ( { model | token = token }, Cmd.none )
                    
        SubmitType type_ ->
            ( { model | type_ = type_ }, Cmd.none )

        SubmitName name ->
            ( { model | name = name }, Cmd.none )

        ProvisionThing ->
            ( model
            , provision
                urls.things
                model.token
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
                urls.things
                model.token
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
                urls.things
                model.name                     
                model.token
            )            



view : Model -> Html Msg
view model =
    Grid.row []
        [ Grid.col []
          [ Form.form []
            [ Form.group []
              [ Form.label [ for "mytype" ] [ text "Type" ]
              , Input.text [ Input.id "mytype", Input.onInput SubmitType ]
              ]
            , Form.group []
                [ Form.label [ for "myname" ] [ text "Name (Provision) or id (Remove)" ]
                , Input.text [ Input.id "myname", Input.onInput SubmitName ]
                ]
            , Form.group []
                [ Form.label [ for "mytoken" ] [ text "Token" ]
                , Input.text [ Input.id "mytoken", Input.onInput SubmitToken ]
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
provision url token type_ name =
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
        , expect = expectProvision ProvisionedThing
        , timeout = Nothing
        , tracker = Nothing
    }


expectProvision : (Result Http.Error Int -> Msg) -> Http.Expect Msg
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


retrieve : String -> String -> Cmd Msg
retrieve url token =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" token ]
        , url = url
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


remove : String -> String -> String -> Cmd Msg
remove url id token =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = url ++ "/" ++ id
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
