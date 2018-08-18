# Mainflux Go SDK

Go SDK, a Go driver for Mainflux HTP API.

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

func ConnectThing(cliId, chanId, token string) (*http.Response, error)
    ConnectThing - connect thing to a channel

func CreateChannel(data, token string) (*http.Response, error)
    CreateChannel - creates new channel and generates UUID

func CreateThing(data, token string) (*http.Response, error)
    CreateThing - creates new thing and generates thing UUID

func CreateToken(user, pwd string) (*http.Response, error)
    CreateToken - create user token

func CreateUser(user, pwd string) (*http.Response, error)
    CreateUser - create user

func DeleteChannel(id, token string) (*http.Response, error)
    DeleteChannel - removes channel

func DeleteThing(id, token string) (*http.Response, error)
    DeleteThing - removes thing

func DisconnectThing(cliId, chanId, token string) (*http.Response, error)
    DisconnectThing - connect thing to a channel

func GetChannel(id, token string) (*http.Response, error)
    GetChannel - gets channel by ID

func GetChannels(token string) (*http.Response, error)
    GetChannels - gets all channels

func GetThing(id, token string) (*http.Response, error)
    GetThing - gets thing by ID

func GetThings(token string) (*http.Response, error)
    GetThings - gets all things

func SendMessage(id, msg, token string) (*http.Response, error)
    SendMessage - send message on Mainflux channel

func SetCerts()
    SetCerts - set TLS certs

func SetContentType(ct string) error
    SetContentType - set message content type. Available options are SenML
    JSON, custom JSON and custom binary (octet-stream).

func SetServerAddr(proto, host string, port int)
    SetServerAddr - set addr using host and port

func UpdateChannel(id, data, token string) (*http.Response, error)
    UpdateChannel - update a channel

func UpdateThing(id, data, token string) (*http.Response, error)
    UpdateThing - updates thing by ID

func Version() (*http.Response, error)
    Version - server health check
```
