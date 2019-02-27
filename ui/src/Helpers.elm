module Helpers exposing (buildQueryParamList, faIcons, fontAwesome, genPagination, pageToOffset, parseName, response, validateInt, validateOffset)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Utilities.Spacing as Spacing
import Html exposing (Html, hr, node, p, text)
import Html.Attributes exposing (..)
import Url.Builder as B


response : String -> Html.Html msg
response resp =
    if String.length resp > 0 then
        Grid.row []
            [ Grid.col []
                [ hr [] []
                , p [] [ text ("response: " ++ resp) ]
                ]
            ]

    else
        Grid.row []
            [ Grid.col [] []
            ]


parseName : Maybe String -> String
parseName thingName =
    case thingName of
        Just name ->
            name

        Nothing ->
            ""



-- PAGINATION


buildQueryParamList : Int -> Int -> List B.QueryParameter
buildQueryParamList offset limit =
    [ B.int "offset" offset, B.int "limit" limit ]


validateInt : String -> Int -> Int
validateInt string default =
    case String.toInt string of
        Just num ->
            num

        Nothing ->
            default


pageToOffset : Int -> Int -> Int
pageToOffset page limit =
    (page - 1) * limit


validateOffset : Int -> Int -> Int -> Int
validateOffset offset total limit =
    if offset >= (total - 1) then
        (total - 1) - limit

    else
        offset


genPagination : Int -> (Int -> msg) -> Html msg
genPagination total msg =
    let
        pages =
            List.range 1 (Basics.ceiling (Basics.toFloat total / 10))

        cols =
            List.map
                (\page ->
                    Grid.col [] [ Button.button [ Button.roleLink, Button.attrs [ Spacing.ml1 ], Button.onClick (msg page) ] [ text (String.fromInt page) ] ]
                )
                pages
    in
    Grid.row [] cols



-- FONT-AWESOME


fontAwesome : Html msg
fontAwesome =
    node "link"
        [ rel "stylesheet"
        , href "https://use.fontawesome.com/releases/v5.7.2/css/all.css"
        , attribute "integrity" "sha384-fnmOCqbTlWIlj8LyTjo7mOUStjsKC4pOpQbqyi7RrhN7udi9RwhKkMHpvLbHG9Sr"
        , attribute "crossorigin" "anonymous"
        ]
        []


faIcons =
    { plus = class "fa fa-plus"
    , pen = class "fa fa-pen"
    , minus = class "fa fa-minus"
    }
