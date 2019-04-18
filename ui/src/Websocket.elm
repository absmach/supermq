-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


module Websocket exposing (Model, Msg(..), initial, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Utilities.Spacing as Spacing
import Debug exposing (log)
import Env exposing (env)
import Error
import Helpers exposing (faIcons, fontAwesome)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http exposing (header)
import HttpMF exposing (paths)
import Json.Decode as D
import Json.Encode as E
import Ports exposing (..)
import Url.Builder as B



-- [{"bn":"some-base-name:","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]


type alias Base =
    { bn : String
    , bt : Float
    , bu : String
    , bv : String
    , bver : Int
    }


type alias Regular =
    { n : String
    , u : String
    , t : Int
    , v : Float
    }


type alias SenML =
    { base : Base
    , regular : Regular
    }


type alias Model =
    { base : Base
    , regular : Regular
    , websocketInMsgs : List String
    }


emptyBase =
    Base "" 0 "" "" 0


emptyRegular =
    Regular "" "" 0 0


initial : Model
initial =
    { base = emptyBase
    , regular = emptyRegular
    , websocketInMsgs = []
    }


type Msg
    = SubmitBaseName String
    | SubmitBaseTime String
    | SubmitBaseUnit String
    | SubmitBaseValue String
    | SubmitBaseVersion String
    | SubmitName String
    | SubmitUnit String
    | SubmitTime String
    | SubmitValue String
    | SendWebsocketMsg
    | ReceiveWebsocketMsg String
    | SubmitReset


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    let
        base =
            model.base

        regular =
            model.regular
    in
    case msg of
        SubmitBaseName bn ->
            ( { model | base = { base | bn = bn } }, Cmd.none )

        SubmitBaseTime bt ->
            let
                btFloat =
                    case String.toFloat bt of
                        Just num ->
                            num

                        Nothing ->
                            0
            in
            ( { model | base = { base | bt = btFloat } }, Cmd.none )

        SubmitBaseUnit bu ->
            ( { model | base = { base | bu = bu } }, Cmd.none )

        SubmitBaseValue bv ->
            ( { model | base = { base | bv = bv } }, Cmd.none )

        SubmitBaseVersion bver ->
            let
                bverInt =
                    case String.toInt bver of
                        Just num ->
                            num

                        Nothing ->
                            0
            in
            ( { model | base = { base | bver = bverInt } }, Cmd.none )

        SubmitName n ->
            ( { model | regular = { regular | n = n } }, Cmd.none )

        SubmitUnit u ->
            ( { model | regular = { regular | u = u } }, Cmd.none )

        SubmitTime t ->
            let
                tInt =
                    case String.toInt t of
                        Just num ->
                            num

                        Nothing ->
                            0
            in
            ( { model | regular = { regular | t = tInt } }, Cmd.none )

        SubmitValue v ->
            let
                vFloat =
                    case String.toFloat v of
                        Just num ->
                            num

                        Nothing ->
                            0
            in
            ( { model | regular = { regular | v = vFloat } }, Cmd.none )

        SendWebsocketMsg ->
            ( model, websocketOut (createSenML model) )

        ReceiveWebsocketMsg inMsg ->
            ( { model | websocketInMsgs = inMsg :: model.websocketInMsgs }
            , Cmd.none
            )

        SubmitReset ->
            ( { model | websocketInMsgs = [] }, Cmd.none )


createSenML : Model -> String
createSenML model =
    E.encode 0 (senMLEncoder model.base model.regular)



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ websocketIn ReceiveWebsocketMsg
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Card.config []
                    |> Card.headerH3 [] [ div [ class "table_header" ] [ i [ style "margin-right" "15px", class faIcons.websocket ] [], text "Base attributes" ] ]
                    |> Card.block []
                        [ Block.custom
                            (Form.form []
                                [ createFormGroup "Base name" SubmitBaseName "prepended to the names"
                                , createFormGroup "Base unit" SubmitBaseUnit "unit assumed for all entries"
                                , createFormGroup "Base version" SubmitBaseVersion "version number of the media type format"
                                , createFormGroup "Base time" SubmitBaseTime "added to the time"
                                , createFormGroup "Base value" SubmitBaseValue "added to the value found in an entry"
                                ]
                            )
                        ]
                    |> Card.view
                ]
            , Grid.col []
                [ Card.config []
                    |> Card.headerH3 [] [ div [ class "table_header" ] [ i [ style "margin-right" "15px", class faIcons.websocket ] [], text "Regular attributes" ] ]
                    |> Card.block []
                        [ Block.custom
                            (Form.form []
                                [ createFormGroup "Name" SubmitName "Name of the sensor or parameter"
                                , createFormGroup "Unit" SubmitUnit "Unit for a measurement value"
                                , createFormGroup "Time" SubmitTime "Time when the value was recorded"
                                , createFormGroup "Value" SubmitValue "Value of the entry"
                                , Button.button
                                    [ Button.success, Button.attrs [ Spacing.ml1 ], Button.onClick SendWebsocketMsg ]
                                    [ text "Send" ]
                                ]
                            )
                        ]
                    |> Card.view
                ]
            ]
        , Grid.row []
            [ Grid.col []
                [ Card.config []
                    |> Card.headerH3 []
                        [ Grid.row []
                            [ Grid.col [ Col.attrs [ align "left" ] ]
                                [ h3 [] [ div [ class "table_header" ] [ i [ style "margin-right" "15px", class faIcons.websocket ] [], text "Received messages" ] ]
                                ]
                            , Grid.col [ Col.attrs [ align "right" ] ]
                                [ Button.button [ Button.secondary, Button.attrs [ align "right" ], Button.onClick SubmitReset ] [ text "Reset" ]
                                ]
                            ]
                        ]
                    |> Card.block []
                        [ Block.custom
                            (model.websocketInMsgs
                                |> List.map li
                                |> Html.ol []
                            )
                        ]
                    |> Card.view
                ]
            ]
        ]


li : String -> Html Msg
li string =
    Html.li [] [ Html.text string ]


createFormGroup : String -> (String -> Msg) -> String -> Html Msg
createFormGroup label msg desc =
    Form.group []
        [ Form.label [] [ text label ]
        , Input.text [ Input.id label, Input.onInput msg ]
        , Form.help [] [ text desc ]
        ]



-- JSON


baseEncoder : Base -> E.Value
baseEncoder base =
    E.object
        [ ( "bn", E.string base.bn )
        , ( "bt", E.float base.bt )
        , ( "bu", E.string base.bu )
        , ( "bv", E.string base.bv )
        , ( "bver", E.int base.bver )
        ]


regularEncoder : Regular -> E.Value
regularEncoder regular =
    E.object
        [ ( "n", E.string regular.n )
        , ( "u", E.string regular.u )
        , ( "t", E.int regular.t )
        , ( "v", E.float regular.v )
        ]


senMLEncoder : Base -> Regular -> E.Value
senMLEncoder base regular =
    E.object
        [ ( "n", baseEncoder base )
        , ( "u", regularEncoder regular )
        ]
