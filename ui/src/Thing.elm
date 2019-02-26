module Thing exposing (Model, Msg(..), Thing, initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.CDN exposing (fontAwesome)
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Debug exposing (log)
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
    , thing : Thing
    , editMode : Bool
    , editName : String
    , modalVisibility : Modal.Visibility
    }


emptyThing =
    Thing "" (Just "") "" ""


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
    , thing = emptyThing
    , editMode = False
    , editName = ""
    , modalVisibility = Modal.hidden
    }


type Msg
    = SubmitType String
    | SubmitName String
    | ProvisionThing
    | ProvisionedThing (Result Http.Error Int)
    | RetrieveThing String
    | RetrievedThing (Result Http.Error Thing)
    | RetrieveThings
    | RetrievedThings (Result Http.Error Things)
    | RemoveThing String
    | RemovedThing (Result Http.Error Int)
    | SubmitPage Int
    | CloseModal
    | ShowModal Thing
    | EditThing
    | EditName String
    | UpdateThing


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
                            | response = String.fromInt statusCode
                            , offset = Helpers.validateOffset model.offset model.things.total query.limit
                            , thing = emptyThing
                            , editMode = False
                            , modalVisibility = Modal.hidden
                        }
                        token

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        CloseModal ->
            ( { model | modalVisibility = Modal.hidden, thing = emptyThing, editMode = False }, Cmd.none )

        ShowModal thing ->
            ( { model | modalVisibility = Modal.shown, thing = thing, editMode = False }, Cmd.none )

        EditThing ->
            ( { model | editMode = True }, Cmd.none )

        EditName name ->
            ( { model | editName = name }, Cmd.none )

        UpdateThing ->
            ( { model | editMode = False, editName = "" }
            , edit
                (B.crossOrigin url.base (List.append url.path [ model.thing.id ]) [])
                token
                model.thing.type_
                model.editName
            )



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ fontAwesome
        , genTable model
        , Helpers.genPagination model.things.total SubmitPage
        , genModal model
        ]



-- Table


genTable : Model -> Html Msg
genTable model =
    Grid.row []
        [ Grid.col []
            [ Table.table
                { options = [ Table.striped, Table.hover ]
                , thead =
                    Table.simpleThead
                        [ Table.th [] [ text "Name" ]
                        , Table.th [] [ text "Id" ]
                        , Table.th [] [ text "Type" ]
                        , Table.th [] [ text "Key" ]
                        ]
                , tbody =
                    Table.tbody []
                        (List.concat
                            [ genTableHeader model.name model.type_
                            , genTableRows model
                            ]
                        )
                }
            ]
        ]


genTableHeader : String -> String -> List (Table.Row Msg)
genTableHeader name type_ =
    [ Table.tr []
        [ Table.td [] [ Input.text [ Input.attrs [ id "name", value name ], Input.onInput SubmitName ] ]
        , Table.td [] []
        , Table.td [] [ Input.text [ Input.attrs [ id "type", value type_ ], Input.onInput SubmitType ] ]
        , Table.td [] []
        , Table.td [] [ Button.button [ Button.outlinePrimary, Button.attrs [ Spacing.ml1, class "fa fa-plus" ], Button.onClick ProvisionThing ] [] ]
        ]
    ]


genTableRows : Model -> List (Table.Row Msg)
genTableRows model =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ text (Helpers.parseName thing.name) ]
                , Table.td [] [ text thing.id ]
                , Table.td [] [ text thing.type_ ]
                , Table.td [] [ text thing.key ]
                , Table.td [] [ Button.button [ Button.outlinePrimary, Button.attrs [ Spacing.ml1, class "fa fa-pencil" ], Button.onClick (ShowModal thing) ] [] ]
                , Table.td [] [ Button.button [ Button.outlineDanger, Button.attrs [ Spacing.ml1, class "fa fa-remove" ], Button.onClick (RemoveThing thing.id) ] [] ]
                ]
        )
        model.things.list



-- Modal


genModal : Model -> Html Msg
genModal model =
    Modal.config CloseModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h4 [] [ text (Helpers.parseName model.thing.name) ]
        |> Modal.body []
            [ Grid.container []
                [ genModalInfo model
                , genModalButtons model
                ]
            ]
        |> Modal.view model.modalVisibility


genModalInfo : Model -> Html Msg
genModalInfo model =
    Grid.row []
        [ Grid.col []
            [ genModalEdit model
            , genModalImmutable model
            ]
        ]


genModalButtons : Model -> Html Msg
genModalButtons model =
    let
        ( msg, buttonText ) =
            if model.editMode then
                ( UpdateThing, "UPDATE" )

            else
                ( EditThing, "EDIT" )
    in
    Grid.row []
        [ Grid.col [ Col.xs8 ]
            [ Button.button [ Button.outlinePrimary, Button.attrs [ Spacing.ml1 ], Button.onClick msg ] [ text buttonText ]
            ]
        , Grid.col []
            [ Button.button [ Button.outlineDanger, Button.attrs [ Spacing.ml1 ], Button.onClick (RemoveThing model.thing.id) ] [ text "REMOVE" ]
            ]
        ]


genModalImmutable : Model -> Html Msg
genModalImmutable model =
    div []
        [ p []
            [ strong [] [ text "type: " ]
            , text model.thing.type_
            ]
        , p []
            [ strong [] [ text "id: " ]
            , text model.thing.id
            ]
        , p []
            [ strong [] [ text "key: " ]
            , text model.thing.key
            ]
        ]


genModalEdit : Model -> Html Msg
genModalEdit model =
    if model.editMode then
        Form.form []
            [ Form.group []
                [ Form.label [] [ strong [] [ text "name" ] ]
                , Input.text [ Input.onInput EditName, Input.attrs [ placeholder (Helpers.parseName model.thing.name) ] ]
                ]
            ]

    else
        div []
            [ p []
                [ strong [] [ text "name: " ]
                , text (Helpers.parseName model.thing.name)
                ]
            ]


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


edit : String -> String -> String -> String -> Cmd Msg
edit u token type_ name =
    Http.request
        { method = "PUT"
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
