module Helpers exposing (buildQueryParamList, genFormField, parseName, response)

import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Html exposing (hr, p, text)
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
