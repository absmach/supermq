module Version exposing (..)

import Http

type Msg
    = GetVersion
    | GotVersion (Result Http.Error String)


