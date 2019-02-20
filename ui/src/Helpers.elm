module Helpers exposing (buildQueryParamList, genFormField, genPagination, parseName, response)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Utilities.Spacing as Spacing
import Html exposing (Html, hr, p, text)
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


buildQueryParamList : String -> String -> { offset : String, limit : String } -> List B.QueryParameter
buildQueryParamList offset limit query =
    List.map
        (\tpl ->
            case String.toInt (Tuple.second tpl) of
                Just n ->
                    B.int (Tuple.first tpl) n

                Nothing ->
                    if Tuple.first tpl == "offset" then
                        B.string (Tuple.first tpl) query.offset

                    else
                        B.string (Tuple.first tpl) query.limit
        )
        [ ( "offset", offset ), ( "limit", limit ) ]


parseName : Maybe String -> String
parseName thingName =
    case thingName of
        Just name ->
            name

        Nothing ->
            ""


genFormField : String -> String -> (String -> msg) -> Grid.Column msg
genFormField txt val msg =
    Grid.col []
        [ Form.formInline []
            [ Form.label [] [ text (txt ++ ": ") ]
            , Input.text [ Input.attrs [ placeholder txt, id txt, value val ], Input.onInput msg ]
            ]
        ]


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
