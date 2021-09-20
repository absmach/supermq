# SMPP Notifier

SMTP Notifier implements notifier for send SMS notifications.

## Configuration

The Subscription service using SMTP Notifier is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                          | Description                                                             | Default               |
| --------------------------------- | ----------------------------------------------------------------------- | --------------------- |
| MF_SMTP_NOTIFY_LOG_LEVEL          | Log level for SMT Notifier (debug, info, warn, error)                   | error                 |
| MF_SMTP_NOTIFY_DB_HOST            | Database host address                                                   | localhost             |
| MF_SMTP_NOTIFY_DB_PORT            | Database host port                                                      | 5432                  |
| MF_SMTP_NOTIFY_DB_USER            | Database user                                                           | mainflux              |
| MF_SMTP_NOTIFY_DB_PASS            | Database password                                                       | mainflux              |
| MF_SMTP_NOTIFY_DB                 | Name of the database used by the service                                | subscriptions         |
| MF_SMTP_NOTIFY_WRITER_CONFIG_PATH | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable               |
| MF_SMTP_NOTIFY_DB_SSL_MODE        | Path to the PEM encoded certificate file                                |                       |
| MF_SMTP_NOTIFY_DB_SSL_CERT        | Path to the PEM encoded key file                                        |                       |
| MF_SMTP_NOTIFY_DB_SSL_KEY         | Path to the PEM encoded root certificate file                           |                       |
| MF_SMTP_NOTIFY_DB_SSL_ROOT_CERT   | Users service HTTP port                                                 | 8180                  |
| MF_SMTP_NOTIFY_HTTP_PORT          | Path to server certificate in pem format                                |                       |
| MF_SMTP_NOTIFY_SERVER_CERT        | Path to server cert in pem format                                       |                       |
| MF_SMTP_NOTIFY_SERVER_KEY         | Path to server key in pem format                                        |                       |
| MF_JAEGER_URL                     | Jaeger server URL                                                       | localhost:6831        |
| MF_NATS_URL                       | NATS broker URL                                                         | nats://127.0.0.1:4222 |
| MF_SMPP_ADDRESS                   | SMPP address                                                            |                       |
| MF_SMPP_USERNAME                  | SMPP Username                                                           |                       |
| MF_SMPP_PASSWORD                  | SMPP PAssword                                                           |                       |
| MF_SMPP_SYSTEM_TYPE               | SMPP System Type                                                        |                       |
| MF_SMPP_SRC_ADDR_TON              | SMPP source address TON                                                 |                       |
| MF_SMPP_DST_ADDR_TON              | SMP destination address TON                                             |                       |
| MF_SMPP_SRC_ADDR_NPI              | SMPP source address NPI                                                 |                       |
| MF_SMPP_DST_ADDR_NPI              | SMPP destination address NPI                                            |                       |
| MF_AUTH_GRPC_TIMEOUT              | Auth service gRPC request timeout in seconds                            | 1s                    |
| MF_AUTH_CLIENT_TLS                | Auth client TLS flag                                                    | false                 |
| MF_AUTH_CA_CERTS                  | Path to Auth client CA certs in pem format                              |                       |

## Usage

Starting service will start consuming messages and sending emails when a message is received.

[doc]: http://mainflux.readthedocs.io
