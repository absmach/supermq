# BOOTSTRAP SERVICE

New devices need to be configured properly and connected to the Mainflux. Bootstrap service is used in order to accomplish that. This service provides the following features:
    1) Creating new Mainflux Thing
    2) Providing basic configuration for the newly created Thing
    3) Handle blacklisting of the Thing

Initial BS endpoint will be provided in the Thing during the manufacturing process. Pre-provisioning a new GW is as simple as sending GW data to the
Bootstrap service. Once the Thing is active it sends a request for initial config to BS service. Once GW is bootstrapped, it’s possible to add it to the whitelist, so that it can exchange messages using Mainflux. Bootstrapping does not implicitly add GW to whitelist, it has to be done manually.

In order to bootstrap successfully, the Thing needs to send bootstrapping request to the specific URL, as well as secret key that are pre-provisioned during manufacturing process. If the Thing is pre-provisioned on the Bootstrap service side, corresponding configuration will be sent as a response. Otherwise, the Thing will be saved so that these Things can be provisioned later.

***Thing Configuration***

Thing Configuration contains two parts: custom configuration (that can be interpreted by the Thing itself) and Mainflux-related configuration. Mainflux config contains:
    1) corresponding Mainflux Thing ID
    2) corresponding Mainflux Thing key
    3) list of the Mainflux channels the Thing is connected to

>Note: list of channels contains IDs of the Mainflux channels. These channels are _pre-provisioned_ on the Mainflux side and Bootstrapping service does not create Mainflux Channels.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                      | Description                                                             | Default               |
|-------------------------------|-------------------------------------------------------------------------|-----------------------|
| MF_BOOTSTRAP_LOG_LEVEL        | Log level for Bootstrap (debug, info, warn, error)                      | error                 |
| MF_BOOTSTRAP_DB_HOST          | Database host address                                                   | localhost             |
| MF_BOOTSTRAP_DB_PORT          | Database host port                                                      | 5432                  |
| MF_BOOTSTRAP_DB_USER          | Database user                                                           | mainflux              |
| MF_BOOTSTRAP_DB_PASS          | Database password                                                       | mainflux              |
| MF_BOOTSTRAP_DB               | Name of the database used by the service                                | things                |
| MF_BOOTSTRAP_DB_SSL_MODE      | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable               |
| MF_BOOTSTRAP_DB_SSL_CERT      | Path to the PEM encoded certificate file                                |                       |
| MF_BOOTSTRAP_DB_SSL_KEY       | Path to the PEM encoded key file                                        |                       |
| MF_BOOTSTRAP_DB_SSL_ROOT_CERT | Path to the PEM encoded root certificate file                           |                       |
| MF_BOOTSTRAP_CLIENT_TLS       | Flag that indicates if TLS should be turned on                          | false                 |
| MF_BOOTSTRAP_CA_CERTS         | Path to trusted CAs in PEM format                                       |                       |
| MF_BOOTSTRAP_PORT             | Bootstrap service HTTP port                                             | 8181                  |
| MF_BOOTSTRAP_SERVER_CERT      | Path to server certificate in pem format                                | 8181                  |
| MF_BOOTSTRAP_SERVER_KEY       | Path to server key in pem format                                        | 8181                  |
| MF_SDK_BASE_URL               | Base url for Mainflux SDK                                               | http://localhost:8182 |
| MF_USERS_URL                  | Users service URL                                                       | localhost:8181        |

## Deployment

The service itself is distributed as Docker container. The following snippet
provides a compose file template that can be used to deploy the service container
locally:

```yaml
version: "2"
  bootstrap:
    image: mainflux/bootstrap:latest
    container_name: nov-bootstrap
    depends_on:
      - bootstrap-db
    restart: on-failure
    ports:
      - 8900:8900
    environment:
      MF_BOOTSTRAP_LOG_LEVEL: [Bootstrap log level]
      MF_BOOTSTRAP_DB_HOST: [Database host address]
      MF_BOOTSTRAP_DB_PORT: [Database host port]
      MF_BOOTSTRAP_DB_USER: [Database user]
      MF_BOOTSTRAP_DB_PASS: [Database password]
      MF_BOOTSTRAP_DB: [Name of the database used by the service]
      MF_BOOTSTRAP_DB_SSL_MODE: [SSL mode to connect to the database with]
      MF_BOOTSTRAP_DB_SSL_CERT: [Path to the PEM encoded certificate file]
      MF_BOOTSTRAP_DB_SSL_KEY: [Path to the PEM encoded key file]
      MF_BOOTSTRAP_DB_SSL_ROOT_CERT: [Path to the PEM encoded root certificate file]
      MF_BOOTSTRAP_CLIENT_TLS: [Boolean value to enable/disable client TLS]
      MF_BOOTSTRAP_CA_CERTS: [Path to trusted CAs in PEM format]
      MF_BOOTSTRAP_PORT: 8900
      MF_BOOTSTRAP_SERVER_CERT: [String path to server cert in pem format]
      MF_BOOTSTRAP_SERVER_KEY: [String path to server key in pem format]
      MF_SDK_BASE_URL: [Base SDK URL for the Mainflux services]
      MF_USERS_URL: [Users service URL]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the things
make things

# copy binary to bin
make install

# set the environment variables and run the service
MF_BOOTSTRAP_LOG_LEVEL=[Bootstrap log level] MF_BOOTSTRAP_DB_HOST=[Database host address] MF_BOOTSTRAP_DB_PORT=[Database host port] MF_BOOTSTRAP_DB_USER=[Database user] MF_BOOTSTRAP_DB_PASS=[Database password] MF_BOOTSTRAP_DB=[Name of the database used by the service] MF_BOOTSTRAP_DB_SSL_MODE=[SSL mode to connect to the database with] MF_BOOTSTRAP_DB_SSL_CERT=[Path to the PEM encoded certificate file] MF_BOOTSTRAP_DB_SSL_KEY=[Path to the PEM encoded key file] MF_BOOTSTRAP_DB_SSL_ROOT_CERT=[Path to the PEM encoded root certificate file] MF_BOOTSTRAP_CLIENT_TLS=[Boolean value to enable/disable client TLS]  MF_BOOTSTRAP_PORT=[Service HTTP port] MF_BOOTSTRAP_SERVER_CERT=[Path to server certificate] MF_BOOTSTRAP_SERVER_KEY=[Path to server key] MF_USERS_URL=[Users service URL]  $GOBIN/mainflux-things
```

Setting `MF_BOOTSTRAP_CA_CERTS` expects a file in PEM format of trusted CAs. This will enable TLS against the Users gRPC endpoint trusting only those CAs that are provided.

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).

[doc]: http://mainflux.readthedocs.io
