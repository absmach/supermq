# Mainflux Go SDK

Go SDK, a Go driver for Mainflux HTTP API.

Does both system administration (provisioning) and messaging.

## Installation
Import `"github.com/mainflux/mainflux/sdk/go"` in your Go package.

```
import "github.com/mainflux/mainflux/sdk/go"
```

Then call SDK Go functions to interact with the system.

## API Reference

```go
FUNCTIONS

func ConnectThing(thingID, chanID, token string) error
    ConnectThing - connect thing to a channel

func CreateChannel(data, token string) (string, error)
    CreateChannel - creates new channel and generates UUID

func CreateThing(data, token string) (string, error)
    CreateThing - creates new thing and generates thing UUID

func CreateToken(user, pwd string) (string, error)
    CreateToken - create user token

func CreateUser(user, pwd string) error
    CreateUser - create user

func DeleteChannel(id, token string) error
    DeleteChannel - removes channel

func DeleteThing(id, token string) error
    DeleteThing - removes thing

func DisconnectThing(thingID, chanID, token string) error
    DisconnectThing - connect thing to a channel

func GetChannel(id, token string) (things.Channel, error)
    GetChannel - gets channel by ID

func GetChannels(token string) ([]things.Channel, error)
    GetChannels - gets all channels

func GetThing(id, token string) (things.Thing, error)
    GetThing - gets thing by ID

func GetThings(token string) ([]things.Thing, error)
    GetThings - gets all things

func SendMessage(id, msg, token string) error
    SendMessage - send message on Mainflux channel

func SetCerts()
    SetCerts - set TLS certs Certs are provided via MF_CERT_FILE,
    MF_KEY_FILE and MF_CA_FILE env vars

func SetContentType(ct string) error
    SetContentType - set message content type. Available options are SenML
    JSON, custom JSON and custom binary (octet-stream).

func SetServerAddr(proto, host, port string)
    SetServerAddr - set addr using host and port

func UpdateChannel(id, data, token string) error
    UpdateChannel - update a channel

func UpdateThing(id, data, token string) error
    UpdateThing - updates thing by ID

func Version() (mainflux.VersionInfo, error)
    Version - server health check
```
