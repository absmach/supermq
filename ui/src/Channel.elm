module Channel exposing (Channel, Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import Url.Builder as B


url =
    { base = "http://localhost"
    , thingsPath = [ "things" ]
    , channelsPath = [ "channels" ]
    }


query =
    { offset = 0
    , limit = 10
    }


type alias Channels =
    { list : List Channel
    , total : Int
    }


type alias Model =
    { name : String
    , offset : Int
    , limit : Int
    , response : String
    , channels : Channels
    }


initial : Model
initial =
    { name = ""
    , offset = query.offset
    , limit = query.limit
    , response = ""
    , channels =
        { list = []
        , total = 0
        }
    }


type Msg
    = SubmitName String
    | ProvisionChannel
    | ProvisionedChannel (Result Http.Error Int)
    | RetrieveChannels
    | RetrieveChannelsForThing String
    | RetrievedChannels (Result Http.Error Channels)
    | RemoveChannel String
    | RemovedChannel (Result Http.Error Int)
    | SubmitPage Int


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitName name ->
            ( { model | name = name }, Cmd.none )

        SubmitPage page ->
            updateChannelList { model | offset = Helpers.pageToOffset page query.limit } token

        ProvisionChannel ->
            ( { model | name = "" }
            , provision
                (B.crossOrigin url.base url.channelsPath [])
                token
                model.name
            )

        ProvisionedChannel result ->
            case result of
                Ok statusCode ->
                    updateChannelList { model | response = String.fromInt statusCode } token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RetrieveChannels ->
            ( model
            , retrieve
                (B.crossOrigin url.base
                    url.channelsPath
                    (Helpers.buildQueryParamList model.offset model.limit)
                )
                token
            )

        RetrieveChannelsForThing thingid ->
            ( model
            , retrieve
                (B.crossOrigin url.base
                    (url.thingsPath ++ [ thingid ] ++ url.channelsPath)
                    (Helpers.buildQueryParamList model.offset model.limit)
                )
                token
            )

        RetrievedChannels result ->
            case result of
                Ok channels ->
                    ( { model | channels = channels }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RemoveChannel id ->
            ( model
            , remove
                (B.crossOrigin url.base (List.append url.channelsPath [ id ]) [])
                token
            )

        RemovedChannel result ->
            case result of
                Ok statusCode ->
                    updateChannelList
                        { model
                            | response = String.fromInt statusCode
                            , offset = Helpers.validateOffset model.offset model.channels.total query.limit
                        }
                        token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Table.simpleTable
                    ( Table.simpleThead
                        [ Table.th [] [ text "Name" ]
                        , Table.th [] [ text "Id" ]
                        ]
                    , Table.tbody []
                        (List.append
                            [ Table.tr []
                                [ Table.td [] [ Input.text [ Input.attrs [ id "name", value model.name ], Input.onInput SubmitName ] ]
                                , Table.td [] []
                                , Table.td [] [ Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick ProvisionChannel ] [ text "+" ] ]
                                ]
                            ]
                            (genTableRows model.channels.list)
                        )
                    )
                ]
            ]
        , Helpers.genPagination model.channels.total SubmitPage
        ]


genTableRows : List Channel -> List (Table.Row Msg)
genTableRows channels =
    List.map
        (\channel ->
            Table.tr []
                [ Table.td [] [ text (Helpers.parseName channel.name) ]
                , Table.td [] [ text channel.id ]
                , Table.td [] [ Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick (RemoveChannel channel.id) ] [ text "-" ] ]
                ]
        )
        channels


type alias Channel =
    { name : Maybe String
    , id : String
    }


channelDecoder : D.Decoder Channel
channelDecoder =
    D.map2 Channel
        (D.maybe (D.field "name" D.string))
        (D.field "id" D.string)


channelsDecoder : D.Decoder Channels
channelsDecoder =
    D.map2 Channels
        (D.field "channels" (D.list channelDecoder))
        (D.field "total" D.int)


provision : String -> String -> String -> Cmd Msg
provision u token name =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body =
            E.object [ ( "name", E.string name ) ]
                |> Http.jsonBody
        , expect = expectStatus ProvisionedChannel
        , timeout = Nothing
        , tracker = Nothing
        }


expectStatus : (Result Http.Error Int -> Msg) -> Http.Expect Msg
expectStatus toMsg =
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


retrieve : String -> String -> Cmd Msg
retrieve u token =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = expectRetrieve RetrievedChannels
        , timeout = Nothing
        , tracker = Nothing
        }


expectRetrieve : (Result Http.Error Channels -> Msg) -> Http.Expect Msg
expectRetrieve toMsg =
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

                Http.GoodStatus_ metadata body ->
                    case D.decodeString channelsDecoder body of
                        Ok value ->
                            Ok value

                        Err err ->
                            Err (Http.BadBody (D.errorToString err))


remove : String -> String -> Cmd Msg
remove u token =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = expectStatus RemovedChannel
        , timeout = Nothing
        , tracker = Nothing
        }



-- HELPERS


updateChannelList : Model -> String -> ( Model, Cmd Msg )
updateChannelList model token =
    ( model
    , retrieve
        (buildUrl url.channelsPath model.offset model.limit)
        token
    )


updateChannelListForThing : Model -> String -> String -> ( Model, Cmd Msg )
updateChannelListForThing model token thingid =
    ( model
    , retrieve
        (buildUrl (url.thingsPath ++ [ thingid ] ++ url.channelsPath) model.offset model.limit)
        token
    )


buildUrl : List String -> Int -> Int -> String
buildUrl path offset limit =
    B.crossOrigin url.base
        path
        (Helpers.buildQueryParamList offset limit)
