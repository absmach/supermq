module Version exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import Url.Builder as B


url =
    { base = "http://localhost"
    , path = [ "version" ]
    }


type alias Model =
    { response : String }


initial : Model
initial =
    { response = "" }


type Msg
    = GetVersion
    | GotVersion (Result Http.Error String)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GetVersion ->
            ( model
            , Http.get
                { url = B.crossOrigin url.base url.path []
                , expect = Http.expectJson GotVersion (D.field "version" D.string)
                }
            )

        GotVersion result ->
            case result of
                Ok text ->
                    ( { model | response = "Version " ++ text }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Button.linkButton
                    [ Button.primary, Button.onClick GetVersion ]
                    [ text "Version" ]
                ]
            ]
        , Helpers.response model.response
        ]
