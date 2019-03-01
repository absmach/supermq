module Helpers exposing (FormRecord, buildQueryParamList, button, editModalButtons, expectStatus, faIcons, fontAwesome, genPagination, modalDiv, modalForm, pageToOffset, parseString, provisionModalButtons, response, validateInt, validateOffset)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Utilities.Spacing as Spacing
import Html exposing (Html, div, hr, node, p, strong, text)
import Html.Attributes exposing (..)
import Http
import Url.Builder as B



-- HTTP


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


expectStatus : (Result Http.Error String -> msg) -> Http.Expect msg
expectStatus toMsg =
    Http.expectStringResponse toMsg <|
        \resp ->
            case resp of
                Http.BadUrl_ u ->
                    Err (Http.BadUrl u)

                Http.Timeout_ ->
                    Err Http.Timeout

                Http.NetworkError_ ->
                    Err Http.NetworkError

                Http.BadStatus_ metadata body ->
                    Err (Http.BadStatus metadata.statusCode)

                Http.GoodStatus_ metadata _ ->
                    Ok (String.fromInt metadata.statusCode)



-- STRING


parseString : Maybe String -> String
parseString str =
    case str of
        Just s ->
            s

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
    { provision = class "fa fa-plus"
    , edit = class "fa fa-pen"
    , remove = class "fa fa-trash-alt"
    }



-- BOOTSTRAP


button type_ msg txt =
    Button.button [ type_, Button.attrs [ Spacing.ml1 ], Button.onClick msg ] [ text txt ]



-- MODAL


modalDiv paragraphList =
    div []
        (List.map
            (\paragraph ->
                p []
                    [ strong [] [ text (Tuple.first paragraph ++ ": ") ]
                    , text (Tuple.second paragraph)
                    ]
            )
            paragraphList
        )


type alias FormRecord msg =
    { text : String
    , msg : String -> msg
    , placeholder : String
    , value : String
    }


modalForm : List (FormRecord msg) -> Html msg
modalForm formList =
    Form.form []
        (List.map
            (\form ->
                Form.group []
                    [ Form.label [] [ strong [] [ text form.text ] ]
                    , Input.text [ Input.onInput form.msg, Input.attrs [ placeholder form.placeholder, value form.value ] ]
                    ]
            )
            formList
        )


editModalButtons mode updateMsg editMsg cancelMsg deleteMsg closeMsg =
    let
        lButton1 =
            if mode then
                button Button.outlinePrimary updateMsg "UPDATE"

            else
                button Button.outlinePrimary editMsg "EDIT"

        lButton2 =
            if mode then
                button Button.outlineDanger cancelMsg "CANCEL"

            else
                button Button.outlineDanger deleteMsg "DELETE"
    in
    Grid.row []
        [ Grid.col [ Col.attrs [ align "left" ] ]
            [ lButton1
            , lButton2
            ]
        , Grid.col [ Col.attrs [ align "right" ] ]
            [ button Button.outlineSecondary closeMsg "CLOSE"
            ]
        ]


provisionModalButtons provisionMsg closeMsg =
    Grid.row []
        [ Grid.col [ Col.attrs [ align "left" ] ]
            [ button Button.outlinePrimary provisionMsg "ADD"
            ]
        , Grid.col [ Col.attrs [ align "right" ] ]
            [ button Button.outlineSecondary closeMsg "CLOSE"
            ]
        ]
