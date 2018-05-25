# Bashflux - Command line interface (CLI) for Mainflux system.
## Quickstart
```
cd bashflux
go build
```

## Usage
### Service
* Get the service verison: `./bashflux version`

### User management
* `./bashflux users create john.doe@email.com password`
* `./bashflux users token john.doe@email.com password`

### System provisioning
* Provisioning devices: `./bashflux things create '{"type":"device", "name":"nyDevice"}' <user_auth_token>`
* Provisioning applications: `./bashflux things create '{"type":"app", "name":"nyDevice"}' <user_auth_token>`
* Retrieving provisioned things: `./bashflux things get --offset=1 --limit=5 <user_auth_token>`
* Retrieving a specific client: `./bashflux things get <client_id>  --offset=1 --limit=5 <user_auth_token>`
* Removing things: ``./bashflux things delete <client_id> <user_auth_token>``

* Provisioning devices: `./bashflux channels create '{"name":"nyChannel"}' <user_auth_token>`
* Provisioning applications: `./bashflux channels create '{"name":"nyChannel"}' <user_auth_token>`
* Retrieving provisioned channels: `./bashflux channels get --offset=1 --limit=5 <user_auth_token>`
* Retrieving a specific channel: `./bashflux channels get <channel_id>  --offset=1 --limit=5 <user_auth_token>`
* Removing channels: `./bashflux channels delete <channel_id> <user_auth_token>`

### Access control
* Connect client to a channel: `./bashflux client connect <client_id> <chanel_id <user_auth_token>`
* Disconnect client from channel: `./bashflux client disconnect <client_id> <chanel_id <user_auth_token>`

* Send message: `./bashflux msg send <channel_id> '[{"bn":"some-base-name:","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]' <client_auth_token>`
