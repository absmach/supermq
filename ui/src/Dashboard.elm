module Dashboard exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Grid as Grid
import Bootstrap.Utilities.Spacing as Spacing
import Debug exposing (log)
import Error
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
    { version : String }


initial : Model
initial =
    { version = "" }


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
                Ok version ->
                    ( { model | version = version }, Cmd.none )

                Err error ->
                    ( { model | version = Error.handle error }, Cmd.none )


view : Model -> Int -> Int -> Html Msg
view model numThings numChannels =
    Grid.container []
        -- [ Card.config [ Card.attrs [ style [ ( "width", "20rem" ) ] ] ]
        [ Card.config []
            |> Card.header [ class "text-center" ]
                [ h3 [ Spacing.mt2 ] [ text "Version" ]
                ]
            |> Card.block []
                [ Block.titleH4 [] [ text model.version ]
                ]
            |> Card.view
        , Card.config []
            |> Card.header [ class "text-center" ]
                [ h3 [ Spacing.mt2 ] [ text "Things" ]
                ]
            |> Card.block []
                [ Block.titleH4 [] [ text (String.fromInt numThings) ]
                ]
            |> Card.view
        , Card.config []
            |> Card.header [ class "text-center" ]
                [ h3 [ Spacing.mt2 ] [ text "Channels" ]
                ]
            |> Card.block []
                [ Block.titleH4 [] [ text (String.fromInt numChannels) ]
                ]
            |> Card.view
        ]
