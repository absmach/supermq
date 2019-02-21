module Connection exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Channel
import Debug exposing (log)
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import List.Extra
import Thing
import Url.Builder as B


url =
    { base = "http://localhost"
    }


type alias Model =
    { thing : String
    , channel : String
    , response : String
    , things : Thing.Model
    , channels : Channel.Model
    , checkedThingsIds : List String
    , checkedChannelsIds : List String
    }


initial : Model
initial =
    { thing = ""
    , channel = ""
    , response = ""
    , things = Thing.initial
    , channels = Channel.initial
    , checkedThingsIds = []
    , checkedChannelsIds = []
    }


type Msg
    = SubmitThing String
    | SubmitChannel String
    | Connect
    | Disconnect
    | ThingMsg Thing.Msg
    | ChannelMsg Channel.Msg
    | GotResponse (Result Http.Error Int)
    | CheckThing String
    | CheckChannel String


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitChannel channel ->
            ( { model | channel = channel }, Cmd.none )

        SubmitThing thing ->
            ( { model | thing = thing }, Cmd.none )

        Connect ->
            if List.isEmpty model.checkedThingsIds || List.isEmpty model.checkedChannelsIds then
                ( model, Cmd.none )

            else
                ( { model | checkedThingsIds = [], checkedChannelsIds = [] }
                , Cmd.batch (modifyConnections model.checkedThingsIds model.checkedChannelsIds "PUT" token)
                )

        Disconnect ->
            if List.isEmpty model.checkedThingsIds || List.isEmpty model.checkedChannelsIds then
                ( model, Cmd.none )

            else
                ( { model | checkedThingsIds = [], checkedChannelsIds = [] }
                , Cmd.batch (modifyConnections model.checkedThingsIds model.checkedChannelsIds "DELETE" token)
                )

        GotResponse result ->
            case result of
                Ok statusCode ->
                    ( { model | response = String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        ThingMsg subMsg ->
            let
                ( updatedThing, thingCmd ) =
                    Thing.update subMsg model.things token
            in
            ( { model | things = updatedThing }, Cmd.map ThingMsg thingCmd )

        ChannelMsg subMsg ->
            let
                ( updatedChannel, channelCmd ) =
                    Channel.update subMsg model.channels token
            in
            ( { model | channels = updatedChannel }, Cmd.map ChannelMsg channelCmd )

        CheckThing id ->
            ( { model | checkedThingsIds = checkEntity id model.checkedThingsIds }, Cmd.none )

        CheckChannel id ->
            ( { model | checkedChannelsIds = checkEntity id model.checkedChannelsIds }, Cmd.none )


checkEntity : String -> List String -> List String
checkEntity id checkedEntitiesIds =
    if List.member id checkedEntitiesIds then
        List.Extra.remove id checkedEntitiesIds

    else
        id :: checkedEntitiesIds


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Html.map ThingMsg
                    (Grid.row []
                        [ Helpers.genFormField "offset" model.things.offset Thing.SubmitOffset
                        , Helpers.genFormField "limit" model.things.limit Thing.SubmitLimit
                        ]
                    )
                , Grid.row []
                    [ Grid.col []
                        [ Table.simpleTable
                            ( Table.simpleThead
                                [ Table.th [] [ text "Name" ]
                                , Table.th [] [ text "Id" ]
                                ]
                            , Table.tbody [] (genThingRows model.checkedThingsIds model.things.things.list)
                            )
                        ]
                    ]
                ]
            , Grid.col []
                [ Html.map ChannelMsg
                    (Grid.row []
                        [ Helpers.genFormField "offset" model.channels.offset Channel.SubmitOffset
                        , Helpers.genFormField "limit" model.channels.limit Channel.SubmitLimit
                        ]
                    )
                , Grid.row []
                    [ Grid.col []
                        [ Table.simpleTable
                            ( Table.simpleThead
                                [ Table.th [] [ text "Name" ]
                                , Table.th [] [ text "Id" ]
                                ]
                            , Table.tbody [] (genChannelRows model.checkedChannelsIds model.channels.channels.list)
                            )
                        ]
                    ]
                ]
            ]
        , Grid.row []
            [ Grid.col []
                [ Form.form []
                    [ Button.button [ Button.success, Button.attrs [ Spacing.ml1 ], Button.onClick Connect ] [ text "Connect" ]
                    , Button.button [ Button.danger, Button.attrs [ Spacing.ml1 ], Button.onClick Disconnect ] [ text "Disonnect" ]
                    ]
                ]
            ]
        , Helpers.response model.response
        ]


genThingRows : List String -> List Thing.Thing -> List (Table.Row Msg)
genThingRows checkedThingsIds things =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ input [ type_ "checkbox", onClick (CheckThing thing.id), checked (isChecked thing.id checkedThingsIds) ] [], text (" " ++ Helpers.parseName thing.name) ]
                , Table.td [] [ text thing.id ]
                ]
        )
        things


genChannelRows : List String -> List Channel.Channel -> List (Table.Row Msg)
genChannelRows checkedChannelsIds channels =
    List.map
        (\channel ->
            Table.tr []
                [ Table.td [] [ input [ type_ "checkbox", onClick (CheckChannel channel.id), checked (isChecked channel.id checkedChannelsIds) ] [], text (" " ++ Helpers.parseName channel.name) ]
                , Table.td [] [ text channel.id ]
                ]
        )
        channels


isChecked : String -> List String -> Bool
isChecked id checkedEntitiesIds =
    if List.member id checkedEntitiesIds then
        True

    else
        False


modifyConnections : List String -> List String -> String -> String -> List (Cmd Msg)
modifyConnections checkedThingsIds checkedChannelsIds method token =
    List.foldr (++)
        []
        (List.map
            (\thingId ->
                List.map
                    (\channelId ->
                        Http.request
                            { method = method
                            , headers = [ Http.header "Authorization" token ]
                            , url = B.crossOrigin url.base [ "channels", channelId, "things", thingId ] []
                            , body = Http.emptyBody
                            , expect = expectResponse GotResponse
                            , timeout = Nothing
                            , tracker = Nothing
                            }
                    )
                    checkedChannelsIds
            )
            checkedThingsIds
        )


expectResponse : (Result Http.Error Int -> Msg) -> Http.Expect Msg
expectResponse toMsg =
    Http.expectStringResponse toMsg <|
        \response ->
            case response of
                Http.BadUrl_ u ->
                    Err (Http.BadUrl u)

                Http.Timeout_ ->
                    Err Http.Timeout

                Http.NetworkError_ ->
                    Err Http.NetworkError

                Http.BadStatus_ metadata body ->
                    Err (Http.BadStatus metadata.statusCode)

                Http.GoodStatus_ metadata _ ->
                    Ok metadata.statusCode
