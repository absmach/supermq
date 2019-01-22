module Helpers exposing (response)

import Bootstrap.Grid as Grid
import Html exposing (hr, p, text)


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
