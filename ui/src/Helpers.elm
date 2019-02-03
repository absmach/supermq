module Helpers exposing (buildQueryParamList, response)

import Bootstrap.Grid as Grid
import Html exposing (hr, p, text)
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
