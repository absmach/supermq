module Thing exposing (Model, Msg(..), Thing, initial, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Debug exposing (log)
import Dict
import Error
import Helpers
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
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


defaultType =
    "app"


type alias Thing =
    { type_ : String
    , name : Maybe String
    , id : String
    , key : String
    , metadata : Maybe String
    }


type alias Things =
    { list : List Thing
    , total : Int
    }


type alias Model =
    { name : String
    , metadata : String
    , type_ : String
    , offset : Int
    , limit : Int
    , response : String
    , things : Things
    , thing : Thing
    , location : String
    , editMode : Bool
    , provisionModalVisibility : Modal.Visibility
    , editModalVisibility : Modal.Visibility
    , provisionDropState : Dropdown.State
    }


emptyThing =
    Thing "" (Just "") "" "" (Just "")


initial : Model
initial =
    { name = ""
    , metadata = ""
    , type_ = defaultType
    , offset = query.offset
    , limit = query.limit
    , response = ""
    , things =
        { list = []
        , total = 0
        }
    , thing = emptyThing
    , location = ""
    , editMode = False
    , provisionModalVisibility = Modal.hidden
    , editModalVisibility = Modal.hidden
    , provisionDropState = Dropdown.initialState
    }


type Msg
    = SubmitType String
    | SubmitName String
    | SubmitMetadata String
    | ProvisionThing
    | ProvisionedThing (Result Http.Error String)
    | EditThing
    | UpdateThing
    | UpdatedThing (Result Http.Error String)
    | RetrieveThing String
    | RetrievedThing (Result Http.Error Thing)
    | RetrieveThings
    | RetrievedThings (Result Http.Error Things)
    | RemoveThing String
    | RemovedThing (Result Http.Error String)
    | SubmitPage Int
    | ClosePorvisionModal
    | CloseEditModal
    | ShowProvisionModal
    | ShowEditModal Thing
    | ProvisionDropState Dropdown.State
    | Type String


update : Msg -> Model -> String -> ( Model, Cmd Msg )
update msg model token =
    case msg of
        SubmitType type_ ->
            ( { model | type_ = type_ }, Cmd.none )

        SubmitName name ->
            ( { model | name = name }, Cmd.none )

        SubmitMetadata metadata ->
            ( { model | metadata = metadata }, Cmd.none )

        SubmitPage page ->
            updateThingList { model | offset = Helpers.pageToOffset page query.limit } token

        ProvisionThing ->
            ( resetEdit model
            , provision
                (B.crossOrigin url.base url.path [])
                token
                model.type_
                model.name
                model.metadata
            )

        ProvisionedThing result ->
            case result of
                Ok thingid ->
                    updateThingList
                        { model
                            | thing = { emptyThing | id = thingid }
                            , provisionModalVisibility = Modal.hidden
                            , editModalVisibility = Modal.shown
                        }
                        token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        EditThing ->
            ( { model
                | editMode = True
                , name = Helpers.parseString model.thing.name
                , metadata = Helpers.parseString model.thing.metadata
              }
            , Cmd.none
            )

        UpdateThing ->
            ( resetEdit { model | editMode = False }
            , updateThing
                (B.crossOrigin url.base (List.append url.path [ model.thing.id ]) [])
                token
                model.thing.type_
                model.name
                model.metadata
            )

        UpdatedThing result ->
            case result of
                Ok statusCode ->
                    updateThingList (resetEdit { model | response = statusCode }) token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RetrieveThing thingid ->
            ( model
            , retrieve
                (B.crossOrigin url.base (List.append url.path [ thingid ]) [])
                token
                RetrievedThing
                thingDecoder
            )

        RetrievedThing result ->
            case result of
                Ok thing ->
                    ( { model | thing = thing }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        RetrieveThings ->
            ( model
            , retrieve
                (B.crossOrigin url.base url.path (Helpers.buildQueryParamList model.offset model.limit))
                token
                RetrievedThings
                thingsDecoder
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
                            | response = statusCode
                            , offset = Helpers.validateOffset model.offset model.things.total query.limit
                            , editModalVisibility = Modal.hidden
                        }
                        token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        ClosePorvisionModal ->
            ( resetEdit { model | provisionModalVisibility = Modal.hidden }, Cmd.none )

        CloseEditModal ->
            ( resetEdit { model | editModalVisibility = Modal.hidden }, Cmd.none )

        ShowProvisionModal ->
            ( { model | provisionModalVisibility = Modal.shown }
            , Cmd.none
            )

        ShowEditModal thing ->
            ( { model
                | editModalVisibility = Modal.shown
                , thing = thing
                , editMode = False
              }
            , Cmd.none
            )

        ProvisionDropState state ->
            ( { model | provisionDropState = state }, Cmd.none )

        Type type_ ->
            ( { model | type_ = type_ }, Cmd.none )



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col [ Col.attrs [ align "right" ] ]
                [ Button.button [ Button.outlinePrimary, Button.attrs [ Spacing.ml1, align "right" ], Button.onClick ShowProvisionModal ] [ text "ADD" ]
                ]
            ]
        , genTable model
        , Helpers.genPagination model.things.total SubmitPage
        , provisionModal model
        , editModal model
        ]



-- Things table


genTable : Model -> Html Msg
genTable model =
    Grid.row []
        [ Grid.col []
            [ Table.table
                { options = [ Table.striped, Table.hover ]
                , thead = genTableHeader
                , tbody = genTableBody model
                }
            ]
        ]


genTableHeader : Table.THead Msg
genTableHeader =
    Table.simpleThead
        [ Table.th [] [ text "Name" ]
        , Table.th [] [ text "Id" ]
        , Table.th [] [ text "Type" ]
        ]


genTableBody : Model -> Table.TBody Msg
genTableBody model =
    Table.tbody []
        (List.map
            (\thing ->
                Table.tr [ Table.rowAttr (onClick (ShowEditModal thing)) ]
                    [ Table.td [] [ text (Helpers.parseString thing.name) ]
                    , Table.td [] [ text thing.id ]
                    , Table.td [] [ text thing.type_ ]
                    ]
            )
            model.things.list
        )



-- Provision modal


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Dropdown.subscriptions model.provisionDropState ProvisionDropState ]


provisionDropDiv : Model -> Html Msg
provisionDropDiv model =
    div []
        [ Dropdown.dropdown
            model.provisionDropState
            { options = []
            , toggleMsg = ProvisionDropState
            , toggleButton =
                Dropdown.toggle [ Button.outlinePrimary ] [ text model.type_ ]
            , items =
                [ Dropdown.buttonItem [ onClick (Type "app") ] [ text "app" ]
                , Dropdown.buttonItem [ onClick (Type "device") ] [ text "device" ]
                ]
            }
        ]


provisionModal : Model -> Html Msg
provisionModal model =
    Modal.config ClosePorvisionModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h4 [] [ text "Add thing" ]
        |> provisionModalBody model
        |> Modal.view model.provisionModalVisibility


provisionModalBody : Model -> (Modal.Config Msg -> Modal.Config Msg)
provisionModalBody model =
    Modal.body []
        [ Grid.container []
            [ Grid.row [] [ Grid.col [] [ provisionDropDiv model ] ]
            , Grid.row [] [ Grid.col [] [ provisionModalForm model ] ]
            , Helpers.provisionModalButtons ProvisionThing ClosePorvisionModal
            ]
        ]


provisionModalForm : Model -> Html Msg
provisionModalForm model =
    Helpers.modalForm
        [ Helpers.FormRecord "name" SubmitName model.name model.name
        , Helpers.FormRecord "metadata" SubmitMetadata model.metadata model.metadata
        ]



-- Edit modal


editModal : Model -> Html Msg
editModal model =
    Modal.config CloseEditModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h4 [] [ text (Helpers.parseString model.thing.name) ]
        |> editModalBody model
        |> Modal.view model.editModalVisibility


editModalBody : Model -> (Modal.Config Msg -> Modal.Config Msg)
editModalBody model =
    Modal.body []
        [ Grid.container []
            [ Grid.row []
                [ Grid.col []
                    [ editModalForm model
                    , Helpers.modalDiv [ ( "type", model.thing.type_ ), ( "id", model.thing.id ), ( "key", model.thing.key ) ]
                    ]
                ]
            , Helpers.editModalButtons model.editMode UpdateThing EditThing (ShowEditModal model.thing) (RemoveThing model.thing.id) CloseEditModal
            ]
        ]


editModalForm : Model -> Html Msg
editModalForm model =
    if model.editMode then
        Helpers.modalForm
            [ Helpers.FormRecord "name" SubmitName (Helpers.parseString model.thing.name) model.name
            , Helpers.FormRecord "metadata" SubmitMetadata (Helpers.parseString model.thing.metadata) model.metadata
            ]

    else
        Helpers.modalDiv [ ( "name", Helpers.parseString model.thing.name ), ( "metadata", Helpers.parseString model.thing.metadata ) ]



-- JSON


thingDecoder : D.Decoder Thing
thingDecoder =
    D.map5 Thing
        (D.field "type" D.string)
        (D.maybe (D.field "name" D.string))
        (D.field "id" D.string)
        (D.field "key" D.string)
        (D.maybe (D.field "metadata" D.string))


thingsDecoder : D.Decoder Things
thingsDecoder =
    D.map2 Things
        (D.field "things" (D.list thingDecoder))
        (D.field "total" D.int)



-- HTTP


expectID : (Result Http.Error String -> Msg) -> Http.Expect Msg
expectID toMsg =
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
                    Ok <|
                        String.dropLeft (String.length "/things/") <|
                            Helpers.parseString (Dict.get "location" metadata.headers)


provision : String -> String -> String -> String -> String -> Cmd Msg
provision u token type_ name metadata =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body =
            E.object
                [ ( "type", E.string type_ )
                , ( "name", E.string name )
                , ( "metadata", E.string metadata )
                ]
                |> Http.jsonBody
        , expect = expectID ProvisionedThing
        , timeout = Nothing
        , tracker = Nothing
        }


updateThing : String -> String -> String -> String -> String -> Cmd Msg
updateThing u token type_ name metadata =
    Http.request
        { method = "PUT"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body =
            E.object
                [ ( "type", E.string type_ )
                , ( "name", E.string name )
                , ( "metadata", E.string metadata )
                ]
                |> Http.jsonBody
        , expect = Helpers.expectStatus UpdatedThing
        , timeout = Nothing
        , tracker = Nothing
        }


expectRetrieve : (Result Http.Error a -> Msg) -> D.Decoder a -> Http.Expect Msg
expectRetrieve toMsg decoder =
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
                    case D.decodeString decoder body of
                        Ok value ->
                            Ok value

                        Err err ->
                            Err (Http.BadBody (D.errorToString err))


retrieve : String -> String -> (Result Http.Error a -> Msg) -> D.Decoder a -> Cmd Msg
retrieve u token msg decoder =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = expectRetrieve msg decoder
        , timeout = Nothing
        , tracker = Nothing
        }


remove : String -> String -> Cmd Msg
remove u token =
    Http.request
        { method = "DELETE"
        , headers = [ Http.header "Authorization" token ]
        , url = u
        , body = Http.emptyBody
        , expect = Helpers.expectStatus RemovedThing
        , timeout = Nothing
        , tracker = Nothing
        }



-- HELPERS


resetEdit : Model -> Model
resetEdit model =
    { model | name = "", type_ = defaultType, metadata = "" }


updateThingList : Model -> String -> ( Model, Cmd Msg )
updateThingList model token =
    ( model
    , Cmd.batch
        [ retrieve
            (B.crossOrigin url.base
                url.path
                (Helpers.buildQueryParamList model.offset model.limit)
            )
            token
            RetrievedThings
            thingsDecoder
        , retrieve
            (B.crossOrigin url.base (List.append url.path [ model.thing.id ]) [])
            token
            RetrievedThing
            thingDecoder
        ]
    )
