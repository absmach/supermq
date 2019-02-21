module Thing exposing (Model, Msg(..), Thing, initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
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


query =
    { offset = 0
    , limit = 10
    }


url =
    { base = "http://localhost"
    , path = [ "things" ]
    }


type alias Things =
    { list : List Thing
    , total : Int
    }


type alias Model =
    { name : String
    , type_ : String
    , offset : Int
    , limit : Int
    , response : String
    , things : Things
    }


initial : Model
initial =
    { name = ""
    , type_ = ""
    , offset = query.offset
    , limit = query.limit
    , response = ""
    , things =
        { list = []
        , total = 0
        }
    }


type Msg
    = SubmitType String
    | SubmitName String
    | ProvisionThing
    | ProvisionedThing (Result Http.Error Int)
    | RetrieveThings
    | RetrievedThings (Result Http.Error Things)
    | RemoveThing String
    | RemovedThing (Result Http.Error Int)
    | SubmitPage Int


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitType type_ ->
            ( { model | type_ = type_ }, Cmd.none )

        SubmitName name ->
            ( { model | name = name }, Cmd.none )

        SubmitPage page ->
            updateThingList { model | offset = Helpers.pageToOffset page query.limit } token

        ProvisionThing ->
            ( { model | name = "", type_ = "" }
            , provision
                (B.crossOrigin url.base url.path [])
                token
                model.type_
                model.name
            )

        ProvisionedThing result ->
            case result of
                Ok statusCode ->
                    updateThingList { model | response = String.fromInt statusCode } token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RetrieveThings ->
            ( model
            , retrieve
                (B.crossOrigin url.base url.path (Helpers.buildQueryParamList model.offset model.limit))
                token
            )

        RetrievedThings result ->
            case result of
                Ok things ->
                    ( { model | things = things }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RemoveThing id ->
            ( model
            , remove
                (B.crossOrigin url.base (List.append url.path [ id ]) [])
                token
            )

        RemovedThing result ->
            case result of
                Ok statusCode ->
                    updateThingList
                        { model
                            | response = String.fromInt statusCode
                            , offset = Helpers.validateOffset model.offset model.things.total query.limit
                        }
                        token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Table.simpleTable
                    ( Table.simpleThead
                        [ Table.th [] [ text "Name" ]
                        , Table.th [] [ text "Id" ]
                        , Table.th [] [ text "Type" ]
                        , Table.th [] [ text "key" ]
                        ]
                    , Table.tbody []
                        (List.concat
                            [ genTableHeader model.name model.type_
                            , genTableRows model.things.list
                            ]
                        )
                    )
                ]
            ]
        , Helpers.genPagination model.things.total SubmitPage
        ]


genTableHeader : String -> String -> List (Table.Row Msg)
genTableHeader name type_ =
    [ Table.tr []
        [ Table.td [] [ Input.text [ Input.attrs [ id "name", value name ], Input.onInput SubmitName ] ]
        , Table.td [] []
        , Table.td [] [ Input.text [ Input.attrs [ id "type", value type_ ], Input.onInput SubmitType ] ]
        , Table.td [] []
        , Table.td [] [ Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick ProvisionThing ] [ text "+" ] ]
        ]
    ]


genTableRows : List Thing -> List (Table.Row Msg)
genTableRows list =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ text (Helpers.parseName thing.name) ]
                , Table.td [] [ text thing.id ]
                , Table.td [] [ text thing.type_ ]
                , Table.td [] [ text thing.key ]
                , Table.td [] [ Button.button [ Button.primary, Button.attrs [ Spacing.ml1 ], Button.onClick (RemoveThing thing.id) ] [ text "-" ] ]
                ]
        )
        list


type alias Thing =
    { type_ : String
    , name : Maybe String
    , id : String
    , key : String
    }


thingDecoder : D.Decoder Thing
thingDecoder =
    D.map4 Thing
        (D.field "type" D.string)
        (D.maybe (D.field "name" D.string))
        (D.field "id" D.string)
        (D.field "key" D.string)


thingsDecoder : D.Decoder Things
thingsDecoder =
    D.map2 Things
        (D.field "things" (D.list thingDecoder))
        (D.field "total" D.int)


provision : String -> String -> String -> String -> Cmd Msg
provision u token type_ name =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body =
            E.object
                [ ( "type", E.string type_ )
                , ( "name", E.string name )
                ]
                |> Http.jsonBody
        , expect = expectStatus ProvisionedThing
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
        , expect = expectRetrieve RetrievedThings
        , timeout = Nothing
        , tracker = Nothing
        }


expectRetrieve : (Result Http.Error Things -> Msg) -> Http.Expect Msg
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
                    case D.decodeString thingsDecoder body of
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
        , expect = expectStatus RemovedThing
        , timeout = Nothing
        , tracker = Nothing
        }


updateThingList : Model -> String -> ( Model, Cmd Msg )
updateThingList model token =
    ( model
    , retrieve
        (B.crossOrigin url.base
            url.path
            (Helpers.buildQueryParamList model.offset model.limit)
        )
        token
    )
