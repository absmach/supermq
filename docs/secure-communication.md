
By default gRPC communication is not secure as Mainflux system is most often run in a private network behind the reverse proxy.

However, TLS can be activated and configured.

# Server configuration

## Securing PostgreSQL connections

By default, Mainflux will connect to Postgres using insecure transport.
If a secured connection is required, you can select the SSL mode and set paths to any extra certificates and keys needed. 

`MF_USERS_DB_SSL_MODE` the SSL connection mode for Users.
`MF_USERS_DB_SSL_CERT` the path to the certificate file for Users.
`MF_USERS_DB_SSL_KEY` the path to the key file for Users.
`MF_USERS_DB_SSL_ROOT_CERT` the path to the root certificate file for Users.

`MF_THINGS_DB_SSL_MODE` the SSL connection mode for Things.
`MF_THINGS_DB_SSL_CERT` the path to the certificate file for Things.
`MF_THINGS_DB_SSL_KEY` the path to the key file for Things.
`MF_THINGS_DB_SSL_ROOT_CERT` the path to the root certificate file for Things.

Supported database connection modes are: `disabled` (default), `required`, `verify-ca` and `verify-full`.

## Securing gRPC

### Users

If either the cert or key is not set, the server will use insecure transport.

`MF_USERS_SERVER_CERT` the path to server certificate in pem format.

`MF_USERS_SERVER_KEY` the path to the server key in pem format.

### Things

If either the cert or key is not set, the server will use insecure transport.

`MF_THINGS_SERVER_CERT` the path to server certificate in pem format.

`MF_THINGS_SERVER_KEY` the path to the server key in pem format.

## Client configuration

If you wish to secure the gRPC connection to `things` and `users` services you must define the CAs that you trust.  This does not support mutual certificate authentication.

### HTTP Adapter

`MF_HTTP_ADAPTER_CA_CERTS` - the path to a file that contains the CAs in PEM format. If not set, the default connection will be insecure. If it fails to read the file, the adapter will fail to start up.

### Things

`MF_THINGS_CA_CERTS` - the path to a file that contains the CAs in PEM format. If not set, the default connection will be insecure. If it fails to read the file, the service will fail to start up.

# Mutual authentication

In the most of the cases, HTTPS, WSS, MQTTS or secure CoAP are secure enough. However, sometimes you might need even more secure connection. Mainflux supports mutual TLS authentication (*mTLS*) based on (X.509 certificates)[https://tools.ietf.org/html/rfc5280]. By default the TLS protocol only proves the identity of the server to the client using X.509 certificate and the authentication of the client to the server is left to the application layer. TLS also offers client-to-server authentication using client-side X.509 authentication. This is called two-way or mutual authentication (_mTLS_). Mainflux currently supports mTLS over HTTP, WS, and MQTT protocols. In order to run Docker composition with mTLS turned on, you can execute following command from the project root:

```bash
AUTH=x509 docker-compose -f docker/docker-compose.yml up -d
```

Mutual authentication includes client side certificates. Certificates can be generated using simple script provided (here)[http://www.github.com/mainflux/mainflux/tree/master/docker/ssl/Makefile]. In order to create a valid certificate, you need to create Mainflux thing using the process described in the [provisioning section](provisioning.md).
