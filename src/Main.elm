module Main exposing (main)

import Debug exposing (log)

import Browser
import Browser.Navigation as Nav
import Html exposing (..)
import Html.Attributes exposing (..)
import Url
import Http
import Json.Decode exposing (Decoder, field, string)

import Bootstrap.CDN as CDN
import Bootstrap.Grid as Grid
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Button as Button


-- MAIN


main : Program () Model Msg
main =
  Browser.application
    { init = init
    , view = view
    , update = update
    , subscriptions = subscriptions
    , onUrlChange = UrlChanged
    , onUrlRequest = LinkClicked
    }



-- MODEL


type alias Model =
  { key : Nav.Key
  , url : Url.Url
  , res : String
  }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
  ( Model key url "", Cmd.none )



-- UPDATE


type Msg
  = LinkClicked Browser.UrlRequest
  | UrlChanged Url.Url
  | GotJSON (Result Http.Error String)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
  case msg of
    LinkClicked urlRequest ->
      case urlRequest of
        Browser.Internal url ->
          ( model, Nav.pushUrl model.key (Url.toString url) )

        Browser.External href ->
            ( model
            , Http.get
                { url = href
                , expect = Http.expectJson GotJSON jsonDecoder
                })

    UrlChanged url ->
      ( { model | url = url }
      , Cmd.none
      )
            
    GotJSON result ->
      case result of
        Ok text ->
            log text
          ({ model | res = text}, Cmd.none)

        Err _ ->
            log "err"
          (model, Cmd.none)


jsonDecoder : Decoder String
jsonDecoder =
    field "version" string


-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
  Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
  { title = "URL Interceptor"
  , body =
      [  Grid.container []
          [ CDN.stylesheet -- creates an inline style node with the Bootstrap CSS, 
          , Button.linkButton
              [ Button.primary, Button.attrs [ href "http://localhost/version" ] ]
              [ text "Version" ]
          , Button.linkButton
              [ Button.secondary, Button.disabled True, Button.attrs [ href "#" ] ]
              [ text "Todo" ]
          , Grid.row []
              [ Grid.col []
                    [ text ("version: " ++ model.res) ]
              ]                                     
          ]
      ]
  }
