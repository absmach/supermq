module User exposing (User, user)

import Http
import Json.Decode exposing (Decoder, field, string)
import Json.Encode as Encode


type alias User =
    { email : String
    , password : String 
    }


email : String -> (String, Encode.Value)
email value =
  ("email", Encode.string value)


password : String -> (String, Encode.Value)
password value =
  ("password", Encode.string value)


user : User -> Encode.Value
user u =
  Encode.object
    [ email u.email
    , password u.password
    ]
