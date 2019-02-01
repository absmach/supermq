# Mainflux CLI
## Build
From the project root:
```
make cli
```

## Usage
### Service
#### Get the service verison
```
mainflux-cli version
```

### Users management
#### Create Users
```
mainflux-cli users create john.doe@email.com password
```

#### Login Users
```
mainflux-cli users token john.doe@email.com password
```

### System Provisioning
#### Create Devices
```
mainflux-cli things create '{"type":"device", "name":"myDevice"}' <user_auth_token>
```

#### Create Applications
```
mainflux-cli things create '{"type":"app", "name":"myDevice"}' <user_auth_token>
```

#### Update Devices and Applications
```
mainflux-cli things update '{"id":"<thing_id>", "name":"myNewName"}' <user_auth_token>
```

#### Remove Things
```
mainflux-cli things delete <thing_id> <user_auth_token>
```

#### Retrieve all provisioned Things
```
mainflux-cli things get all --offset=1 --limit=5 <user_auth_token>
```

#### Retrieve Things By ID
```
mainflux-cli things get <thing_id> <user_auth_token>
```

#### Create Channels
```
mainflux-cli channels create '{"name":"myChannel"}' <user_auth_token>
```

#### Update Channels
```
mainflux-cli channels update '{"id":"<channel_id>","name":"myNewName"}' <user_auth_token>

```
#### Remove Channels
```
mainflux-cli channels delete <channel_id> <user_auth_token>
```

#### Retrieve all provisioned Channels
```
mainflux-cli channels get all --offset=1 --limit=5 <user_auth_token>
```

#### Retrieve Channels By ID
```
mainflux-cli channels get <channel_id> <user_auth_token>
```

### Access control
#### Connect Things to Channels
```
mainflux-cli things connect <thing_id> <channel_id> <user_auth_token>
```

#### Disconnect Things from Channels
```
mainflux-cli things disconnect <thing_id> <channel_id> <user_auth_token>

```

#### Retrieve List of Channels connected to Things
```
mainflux-cli things connections <thing_id> <user_auth_token>
```

#### Retrieve List of Things connected to Channels
```
mainflux-cli channels connections <channel_id> <user_auth_token>
```

### Messaging
#### Send a message over HTTP
```
mainflux-cli msg send <channel_id> '[{"bn":"Dev1","n":"temp","v":20}, {"n":"hum","v":40}, {"bn":"Dev2", "n":"temp","v":20}, {"n":"hum","v":40}]' <thing_auth_token>
```
