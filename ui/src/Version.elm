module Version exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Grid as Grid
import Bootstrap.Text as Text
import Bootstrap.Utilities.Spacing as Spacing
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import Thing
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
    -- Grid.container []
    --     [ Grid.row []
    --         [ Grid.col [] [ text model.response ]
    --         ]
    --     ]
    -- Grid.container []
    --     [ Grid.row []
    --         [ Grid.col []
    --             [ Card.config
    --                 []
    --                 |> Card.header [ class "text-center" ]
    --                     [ img [ src "assets/images/elm-bootstrap.svg" ] []
    --                     , h3 [ Spacing.mt2 ] [ text "Custom Card Header" ]
    --                     ]
    --                 |> Card.block []
    --                     [ Block.titleH4 [] [ text "Card title" ]
    --                     , Block.text [] [ text "Some quick example text to build on the card title and make up the bulk of the card's content." ]
    --                     , Block.custom <|
    --                         Button.button [ Button.primary ] [ text "Go somewhere" ]
    --                     ]
    --                 |> Card.view
    --             ]
    --         ]
    --     ]
    Grid.container
        []
        [ Grid.row []
            [ Grid.col []
                [ Card.deck (cardList model)
                ]
            ]
        ]



-- Card.config
--     [ Card.attrs [] ]
--     |> Card.header [ class "text-center" ]
--         [ img [ src "assets/images/elm-bootstrap.svg" ] []
--         , h3 [ Spacing.mt2 ] [ text "Custom Card Header" ]
--         ]
--     |> Card.block []
--         [ Block.titleH4 [] [ text "Card title" ]
--         , Block.text [] [ text "Some quick example text to build on the card title and make up the bulk of the card's content." ]
--         , Block.custom <|
--             Button.button [ Button.primary ] [ text "Go somewhere" ]
--         ]
--     |> Card.view


cardList : Model -> List (Card.Config msg)
cardList model =
    [ Card.config
        [ Card.secondary
        , Card.textColor Text.white
        ]
        |> Card.headerH3 [] [ text "Version" ]
        |> Card.block []
            [ Block.titleH4 [] [ text model.response ] ]
    , Card.config
        [ Card.info
        , Card.textColor Text.white
        ]
        |> Card.headerH3 [] [ text "Things" ]
        |> Card.block []
            [ Block.titleH4 [] [ text "69" ]
            , Block.custom <|
                Button.button [ Button.light ] [ text "Manage things" ]
            ]
    , Card.config []
        |> Card.headerH3 [] [ text "Channels" ]
        |> Card.block []
            [ Block.titleH4 [] [ text "69" ]
            , Block.custom <|
                Button.button [ Button.dark ] [ text "Manage channels" ]
            ]
    ]
