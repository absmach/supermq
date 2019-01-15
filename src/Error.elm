module Error exposing (..)


import Http


handle : Http.Error -> String
handle error =
    case error of
        Http.BadUrl txt ->
            "Bad url " ++ txt

        Http.Timeout ->
            "Timeout"

        Http.NetworkError ->
            "Network error"

        Http.BadStatus num ->
            "Bad status " ++ String.fromInt num

        Http.BadBody err ->
            err

