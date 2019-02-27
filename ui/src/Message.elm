module Message exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Radio as Radio
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Channel
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
    , httpPath = [ "http" ]
    , thingsPath = [ "things" ]
    , channelsPath = [ "channels" ]
    , messagesPath = [ "messages" ]
    }


type alias Model =
    { message : String
    , thingkey : String
    , response : String
    , things : Thing.Model
    , channels : Channel.Model
    , thingid : String
    , checkedChannelsIds : List String
    }


initial : Model
initial =
    { message = ""
    , thingkey = ""
    , response = ""
    , things = Thing.initial
    , channels = Channel.initial
    , thingid = ""
    , checkedChannelsIds = []
    }


type Msg
    = SubmitMessage String
    | SendMessage
    | SentMessage (Result Http.Error Int)
    | ThingMsg Thing.Msg
    | ChannelMsg Channel.Msg
    | SelectedThing String String Channel.Msg
    | CheckChannel String


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitMessage message ->
            ( { model | message = message }, Cmd.none )

        SendMessage ->
            ( { model | message = "", thingkey = "", response = "", thingid = "" }
            , Cmd.batch
                (List.map
                    (\channelId ->
                        Http.request
                            { method = "POST"
                            , headers = [ Http.header "Authorization" model.thingkey ]
                            , url = B.crossOrigin url.base (url.httpPath ++ url.channelsPath ++ [ channelId ] ++ url.messagesPath) []
                            , body = Http.stringBody "application/json" model.message
                            , expect = expectMessage SentMessage
                            , timeout = Nothing
                            , tracker = Nothing
                            }
                    )
                    model.checkedChannelsIds
                )
            )

        SentMessage result ->
            case result of
                Ok statusCode ->
                    ( { model | response = String.fromInt statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        ThingMsg subMsg ->
            updateThing model subMsg token

        ChannelMsg subMsg ->
            updateChannel model subMsg token

        SelectedThing thingid thingkey channelMsg ->
            updateChannel { model | thingid = thingid, thingkey = thingkey, checkedChannelsIds = [] } (Channel.RetrieveChannelsForThing thingid) token

        CheckChannel id ->
            ( { model | checkedChannelsIds = checkEntity id model.checkedChannelsIds }, Cmd.none )


updateThing : Model -> Thing.Msg -> String -> ( Model, Cmd Msg )
updateThing model msg token =
    let
        ( updatedThing, thingCmd ) =
            Thing.update msg model.things token
    in
    ( { model | things = updatedThing }, Cmd.map ThingMsg thingCmd )


updateChannel : Model -> Channel.Msg -> String -> ( Model, Cmd Msg )
updateChannel model msg token =
    let
        ( updatedChannel, channelCmd ) =
            Channel.update msg model.channels token
    in
    ( { model | channels = updatedChannel }, Cmd.map ChannelMsg channelCmd )



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Grid.row []
                    [ Grid.col []
                        [ Table.simpleTable
                            ( Table.simpleThead
                                [ Table.th [] [ text "Name" ]
                                , Table.th [] [ text "Id" ]
                                ]
                            , Table.tbody [] (genThingRows model.things.things.list)
                            )
                        ]
                    ]
                , Html.map ThingMsg (Helpers.genPagination model.things.things.total Thing.SubmitPage)
                ]
            , Grid.col []
                [ Grid.row []
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
                , Html.map ChannelMsg (Helpers.genPagination model.channels.channels.total Channel.SubmitPage)
                ]
            ]
        , Grid.row []
            [ Grid.col []
                [ Form.form []
                    [ Form.group []
                        [ Form.label [ for "message" ] [ text "Message" ]
                        , Input.text [ Input.id "message", Input.onInput SubmitMessage ]
                        ]
                    , Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick SendMessage ] [ text "Send" ]
                    ]
                ]
            ]
        , Helpers.response model.response
        ]


genThingRows : List Thing.Thing -> List (Table.Row Msg)
genThingRows things =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ label [] [ input [ type_ "radio", onClick (SelectedThing thing.id thing.key (Channel.RetrieveChannelsForThing thing.id)), name "things" ] [], text (Helpers.parseName thing.name) ] ]
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


checkEntity : String -> List String -> List String
checkEntity id checkedEntitiesIds =
    if List.member id checkedEntitiesIds then
        List.Extra.remove id checkedEntitiesIds

    else
        id :: checkedEntitiesIds


expectMessage : (Result Http.Error Int -> msg) -> Http.Expect msg
expectMessage toMsg =
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
