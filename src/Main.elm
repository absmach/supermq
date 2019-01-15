module Main exposing (Model, Msg(..), init, main, menuButtons, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.CDN as CDN
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Fieldset as Fieldset
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Radio as Radio
import Bootstrap.Form.Select as Select
import Bootstrap.Form.Textarea as Textarea
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Utilities.Spacing as Spacing
import Browser
import Browser.Navigation as Nav
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode exposing (Decoder, field, string)
import Json.Encode as Encode
import Url
import Url.Parser as UrlParser exposing ((</>))


import Error
import User
import Thing
import Channel


urls =
    { version = "http://localhost/version"
    , things = "http://localhost/things"
    }


-- MAIN


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }


-- MODEL


type alias Model =
    { key : Nav.Key
    , route : Maybe DocsRoute
    , response : String
    , email : String
    , password : String
    , token : String
    , thingType : String
    , thingName : String
    , user : User.Model
    , channel : Channel.Model
    }
    

init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( Model key (UrlParser.parse docsParser url)
          "" "" "" "" "" ""
          User.initial
          Channel.initial
    , Cmd.none )


-- UPDATE


type alias DocsRoute =
    ( String, Maybe String )


docsParser : UrlParser.Parser (DocsRoute -> a) a
docsParser =
    UrlParser.map Tuple.pair (UrlParser.string </> UrlParser.fragment identity)


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | GetVersion
    | GotVersion (Result Http.Error String)
    | UserMsg User.Msg
    | ChannelMsg Channel.Msg
    | SubmitToken String      
    | SubmitThingType String
    | SubmitThingName String
    | ProvisionThing
    | ProvisionedThing (Result Http.Error Int)      
    | RetrieveThing
    | RetrievedThing (Result Http.Error (List Thing.Thing))
    | RemoveThing


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( Debug.log "model before click:" model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( Debug.log "model EXTERNAL click:" model, Cmd.none )

        UrlChanged url ->
            ( { model | route = UrlParser.parse docsParser url }
            , Cmd.none
            )

        GetVersion ->
            ( model
            , Http.get
                { url = urls.version
                , expect = Http.expectJson GotVersion (field "version" string)
                }
            )

        GotVersion result ->
            case result of
                Ok text ->
                    ( Debug.log "model GOT version OK: " { model | response = "Version " ++ text }, Cmd.none )

                Err _ ->
                    ( Debug.log "model GOT version ERR: " model, Cmd.none )

        UserMsg subMsg ->
           let
                ( updatedUser, userCmd ) =
                    User.update subMsg model.user
            in
                ( { model | user = updatedUser }, Cmd.map UserMsg userCmd )

        ChannelMsg subMsg ->
           let
                ( updatedChannel, channelCmd ) =
                    Channel.update subMsg model.channel
            in
                ( { model | channel = updatedChannel }, Cmd.map ChannelMsg channelCmd )

        SubmitToken token ->
            ( { model | token = token }, Cmd.none )
                    
        SubmitThingType type_ ->
            ( { model | thingType = type_ }, Cmd.none )

        SubmitThingName name ->
            ( { model | thingName = name }, Cmd.none )

        ProvisionThing ->
            ( model
            , Thing.provision
                urls.things
                model.token
                model.thingType
                model.thingName
                (Thing.expectProvision ProvisionedThing)
            )

        ProvisionedThing result ->
            case result of
                Ok statusCode ->
                    ( { model | response = "Ok " ++ String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )
            
        RetrieveThing ->
            ( model
            , Thing.retrieve
                urls.things
                model.token
                (Thing.expectRetrieve RetrievedThing)
            )
            

        RetrievedThing result ->
            case result of
                Ok things ->
                    ( { model | response = thingsToString things }, Cmd.none )
                    
                Err error ->
                    ( { model | response = (Error.handle error) }, Cmd.none )
            
        RemoveThing ->
            ( model
            , Thing.remove
                urls.things
                model.thingName                     
                model.token
                (Thing.expectProvision ProvisionedThing)                     
            )            

                
-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Gateflux"
    , body =
        let
            response =
                Grid.row []
                    [ Grid.col []
                        [ text ("response: " ++ model.response) ]
                    ]

            content =
                case model.route of
                    Just route ->
                        case Tuple.first route of
                            "version" ->
                                [ Button.linkButton
                                    [ Button.primary, Button.onClick GetVersion ]
                                    [ text "Version" ]
                                , hr [] []
                                , response
                                ]

                            "account" ->
                                [ Html.map UserMsg (User.view model.user) ]
                                    
                            "channel" ->
                                [ Html.map ChannelMsg (Channel.view model.channel) ]                                

                            "things" ->
                                [ Form.form []
                                    [ Form.group []
                                        [ Form.label [ for "mytype" ] [ text "Type" ]
                                        , Input.text [ Input.id "mytype", Input.onInput SubmitThingType ]
                                        ]
                                    , Form.group []
                                        [ Form.label [ for "myname" ] [ text "Name (Provision) | Id (Retrieve)" ]
                                        , Input.text [ Input.id "myname", Input.onInput SubmitThingName ]
                                        ]
                                    , Form.group []
                                        [ Form.label [ for "mytoken" ] [ text "Token" ]
                                        , Input.text [ Input.id "mytoken", Input.onInput SubmitToken ]
                                        ]                                        
                                    , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick ProvisionThing ] [ text "Provision" ]
                                    , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RetrieveThing ] [ text "Retrieve" ]
                                    , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick RemoveThing ] [ text "Remove" ]
                                    ]
                                , hr [] []
                                , response
                                ]

                            _ ->
                                [ h3 [] [ text "Welcome to Gateflux" ] ]

                    Nothing ->
                        [ h3 [] [ text "Welcome" ] ]
        in
        [ -- we use Bootstrap container defined at http://elm-bootstrap.info/grid
          Grid.container []
            [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS
            , Grid.row []
                [ Grid.col [] [ h1 [] [ text "Gateflux" ] ] ]
            , Grid.row []
                [ Grid.col []
                    [ -- In this column we put the button group defined below
                      ButtonGroup.linkButtonGroup [ ButtonGroup.vertical ] menuButtons
                    ]
                , Grid.col [ Col.xs10 ]
                    [ div
                        []
                        content
                    ]
                ]
            ]
        ]
    }



-- Vertical button group defined at http://elm-bootstrap.info/buttongroup


menuButtons : List (ButtonGroup.LinkButtonItem msg)
menuButtons =
    [ ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/version" ] ] [ text "Version" ]
    , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/account" ] ] [ text "Account" ]
    , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/channel" ] ] [ text "Channel" ]
    , ButtonGroup.linkButton [ Button.secondary, Button.attrs [ href "/things" ] ] [ text "Things" ]
    ]


-- HELPERS


thingsToString : List Thing.Thing -> String
thingsToString things =
    List.map
        (\thing -> thing.id ++ " " ++ thing.type_ ++ " " ++ thing.name ++ " " ++ thing.key ++ "; ")
        things
        |> String.concat
