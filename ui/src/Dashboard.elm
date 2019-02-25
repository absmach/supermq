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
    , versionPath = [ "version" ]
    }


type alias Model =
    {}


initial : Model
initial =
    {}


type Msg
    = NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )


view : Model -> String -> Int -> Int -> Html Msg
view model version numThings numChannels =
    Grid.container []
        -- [ Card.config [ Card.attrs [ style [ ( "width", "20rem" ) ] ] ]
        [ Card.config []
            |> Card.header [ class "text-center" ]
                [ h3 [ Spacing.mt2 ] [ text "Version" ]
                ]
            |> Card.block []
                [ Block.titleH4 [] [ text version ]
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
