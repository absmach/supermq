# Groups

The Groups service exposes HTTP and gRPC APIs for organizing entities into hierarchical groups within a domain, managing membership, permissions, and roles. It handles group lifecycle (create/update/enable/disable/delete), parent/child relationships, listings (flat or tree), and role-based access.

For a deeper overview of SuperMQ, see the [official documentation][doc].

## Configuration

The service is configured via environment variables (unset values fall back to defaults).

| Variable                               | Description                                                                                       | Default                               |
| -------------------------------------- | ------------------------------------------------------------------------------------------------- | ------------------------------------- |
| `SMQ_GROUPS_LOG_LEVEL`                 | Log level for Groups (debug, info, warn, error)                                                   | debug                                  |
| `SMQ_GROUPS_HTTP_HOST`                 | Groups service HTTP host                                                                          | groups                                 |
| `SMQ_GROUPS_HTTP_PORT`                 | Groups service HTTP port                                                                          | 9004                                   |
| `SMQ_GROUPS_HTTP_SERVER_CERT`          | Path to PEM-encoded HTTP server certificate                                                       | ""                                     |
| `SMQ_GROUPS_HTTP_SERVER_KEY`           | Path to PEM-encoded HTTP server key                                                               | ""                                     |
| `SMQ_GROUPS_HTTP_SERVER_CA_CERTS`      | Path to trusted CA bundle for the HTTP server                                                     | ""                                     |
| `SMQ_GROUPS_HTTP_CLIENT_CA_CERTS`      | Path to client CA bundle to require HTTP mTLS                                                     | ""                                     |
| `SMQ_GROUPS_GRPC_HOST`                 | Groups service gRPC host                                                                          | groups                                 |
| `SMQ_GROUPS_GRPC_PORT`                 | Groups service gRPC port                                                                          | 7004                                   |
| `SMQ_GROUPS_GRPC_SERVER_CERT`          | Path to PEM-encoded gRPC server certificate                                                       | ""                                     |
| `SMQ_GROUPS_GRPC_SERVER_KEY`           | Path to PEM-encoded gRPC server key                                                               | ""                                     |
| `SMQ_GROUPS_GRPC_SERVER_CA_CERTS`      | Path to trusted CA bundle for the gRPC server                                                     | ""                                     |
| `SMQ_GROUPS_GRPC_CLIENT_CA_CERTS`      | Path to client CA bundle to require gRPC mTLS                                                     | ""                                     |
| `SMQ_GROUPS_DB_HOST`                   | Database host address                                                                             | groups-db                              |
| `SMQ_GROUPS_DB_PORT`                   | Database host port                                                                                | 5432                                   |
| `SMQ_GROUPS_DB_USER`                   | Database user                                                                                     | supermq                                |
| `SMQ_GROUPS_DB_PASS`                   | Database password                                                                                 | supermq                                |
| `SMQ_GROUPS_DB_NAME`                   | Name of the database used by the service                                                          | groups                                 |
| `SMQ_GROUPS_DB_SSL_MODE`               | Database connection SSL mode (disable, require, verify-ca, verify-full)                           | disable                                |
| `SMQ_GROUPS_DB_SSL_CERT`               | Path to the PEM-encoded certificate file                                                          | ""                                     |
| `SMQ_GROUPS_DB_SSL_KEY`                | Path to the PEM-encoded key file                                                                  | ""                                     |
| `SMQ_GROUPS_DB_SSL_ROOT_CERT`          | Path to the PEM-encoded root certificate file                                                     | ""                                     |
| `SMQ_GROUPS_INSTANCE_ID`               | Groups instance ID (auto-generated when empty)                                                    | ""                                     |
| `SMQ_GROUPS_EVENT_CONSUMER`            | NATS consumer name for domain events                                                              | groups                                 |
| `SMQ_SPICEDB_HOST`                     | SpiceDB host for policy checks                                                                    | supermq-spicedb                              |
| `SMQ_SPICEDB_PORT`                     | SpiceDB port                                                                                      | 50051                                  |
| `SMQ_SPICEDB_SCHEMA_FILE`              | Path to SpiceDB schema file used to seed available actions                                        | "/schema.zed"                              |
| `SMQ_SPICEDB_PRE_SHARED_KEY`           | SpiceDB preshared key                                                                             | 12345678                               |
| `SMQ_ES_URL`                           | Event store URL                                                                                   | nats://nats:4222                  |
| `SMQ_JAEGER_URL`                       | Jaeger server URL                                                                                 | <http://jaeger:4318/v1/traces>      |
| `SMQ_JAEGER_TRACE_RATIO`               | Trace sampling ratio                                                                              | 1.0                                    |
| `SMQ_SEND_TELEMETRY`                   | Send telemetry to the SuperMQ call-home server                                                    | true                                   |
| `SMQ_AUTH_GRPC_URL`                    | Auth service gRPC URL                                                                             | ""                                     |
| `SMQ_AUTH_GRPC_TIMEOUT`                | Auth service gRPC request timeout                                                                 | 1s                                     |
| `SMQ_AUTH_GRPC_CLIENT_CERT`            | Path to the PEM-encoded Auth gRPC client certificate                                              | ""                                     |
| `SMQ_AUTH_GRPC_CLIENT_KEY`             | Path to the PEM-encoded Auth gRPC client key                                                      | ""                                     |
| `SMQ_AUTH_GRPC_SERVER_CA_CERTS`        | Path to the PEM-encoded Auth gRPC trusted CA bundle                                               | ""                                     |
| `SMQ_GROUPS_CALLOUT_URLS`              | Comma-separated list of HTTP callout targets invoked on group operations                          | ""                                     |
| `SMQ_GROUPS_CALLOUT_METHOD`            | HTTP method for callouts (POST or GET)                                                            | POST                                   |
| `SMQ_GROUPS_CALLOUT_TLS_VERIFICATION`  | Verify TLS certificates for callouts                                                              | false                                  |
| `SMQ_GROUPS_CALLOUT_TIMEOUT`           | Callout request timeout                                                                           | 10s                                    |
| `SMQ_GROUPS_CALLOUT_CA_CERT`           | CA bundle for verifying callout targets                                                           | ""                                     |
| `SMQ_GROUPS_CALLOUT_CERT`              | Client certificate for mTLS callouts                                                              | ""                                     |
| `SMQ_GROUPS_CALLOUT_KEY`               | Client key for mTLS callouts                                                                      | ""                                     |
| `SMQ_GROUPS_CALLOUT_OPERATIONS`        | Comma-separated list of operation names that should trigger callouts                              | ""                                     |

**Note**: Set `SMQ_GROUPS_CALLOUT_OPERATIONS` to a subset of `OpCreateGroup`, `OpViewGroup`, `OpUpdateGroup`, `OpEnableGroup`, `OpDisableGroup`, `OpDeleteGroup`, `OpListGroups`, `OpHierarchy`, `OpAddParentGroup`, `OpRemoveParentGroup`, `OpAddChildrenGroups`, `OpRemoveChildrenGroups`, `OpRemoveAllChildrenGroups`, or `OpListChildrenGroups` to filter which actions produce callouts.

## Deployment

The service ships as a Docker container. See the [`groups` section](https://github.com/absmach/supermq/blob/main/docker/docker-compose.yaml#L950-L1035) in `docker-compose.yaml` for deployment configuration.

To build and run locally:

```bash
# download the latest version of the service
git clone https://github.com/absmach/supermq
cd supermq

# compile the groups service
make groups

# copy binary to $GOBIN
make install

# set the environment variables and run the service
SMQ_GROUPS_LOG_LEVEL=debug \
SMQ_GROUPS_HTTP_HOST=groups \
SMQ_GROUPS_HTTP_PORT=9004 \
SMQ_GROUPS_HTTP_SERVER_CERT="" \
SMQ_GROUPS_HTTP_SERVER_KEY="" \
SMQ_GROUPS_GRPC_HOST=groups \
SMQ_GROUPS_GRPC_PORT=7004 \
SMQ_GROUPS_GRPC_SERVER_CERT="" \
SMQ_GROUPS_GRPC_SERVER_KEY="" \
SMQ_GROUPS_GRPC_SERVER_CA_CERTS="" \
SMQ_GROUPS_GRPC_CLIENT_CA_CERTS="" \
SMQ_GROUPS_DB_HOST=groups-db \
SMQ_GROUPS_DB_PORT=5432 \
SMQ_GROUPS_DB_USER=supermq \
SMQ_GROUPS_DB_PASS=supermq \
SMQ_GROUPS_DB_NAME=groups \
SMQ_GROUPS_DB_SSL_MODE=disable \
SMQ_GROUPS_DB_SSL_CERT="" \
SMQ_GROUPS_DB_SSL_KEY="" \
SMQ_GROUPS_DB_SSL_ROOT_CERT="" \
SMQ_AUTH_GRPC_URL="" \
SMQ_AUTH_GRPC_TIMEOUT=1s \
SMQ_AUTH_GRPC_CLIENT_CERT="" \
SMQ_AUTH_GRPC_CLIENT_KEY="" \
SMQ_AUTH_GRPC_SERVER_CA_CERTS="" \
SMQ_DOMAINS_GRPC_URL=domains:7003 \
SMQ_DOMAINS_GRPC_TIMEOUT=1s \
SMQ_DOMAINS_GRPC_CLIENT_CERT="" \
SMQ_DOMAINS_GRPC_CLIENT_KEY="" \
SMQ_DOMAINS_GRPC_SERVER_CA_CERTS="" \
SMQ_CHANNELS_GRPC_URL=channels:7005 \
SMQ_CHANNELS_GRPC_TIMEOUT=1s \
SMQ_CHANNELS_GRPC_CLIENT_CERT="" \
SMQ_CHANNELS_GRPC_CLIENT_KEY="" \
SMQ_CHANNELS_GRPC_SERVER_CA_CERTS="" \
SMQ_CLIENTS_GRPC_URL=clients:7000 \
SMQ_CLIENTS_GRPC_TIMEOUT=1s \
SMQ_CLIENTS_GRPC_CLIENT_CERT="" \
SMQ_CLIENTS_GRPC_CLIENT_KEY="" \
SMQ_CLIENTS_GRPC_SERVER_CA_CERTS="" \
SMQ_SPICEDB_HOST=localhost \
SMQ_SPICEDB_PORT=50051 \
SMQ_SPICEDB_SCHEMA_FILE=schema.zed \
SMQ_SPICEDB_PRE_SHARED_KEY=12345678 \
SMQ_ES_URL=nats://localhost:4222 \
SMQ_JAEGER_URL=<http://localhost:4318/v1/traces> \
SMQ_JAEGER_TRACE_RATIO=1.0 \
SMQ_GROUPS_CALLOUT_URLS="" \
SMQ_GROUPS_CALLOUT_METHOD=POST \
SMQ_GROUPS_CALLOUT_TLS_VERIFICATION=false \
SMQ_GROUPS_CALLOUT_TIMEOUT=10s \
SMQ_GROUPS_CALLOUT_CA_CERT="" \
SMQ_GROUPS_CALLOUT_CERT="" \
SMQ_GROUPS_CALLOUT_KEY="" \
SMQ_GROUPS_CALLOUT_OPERATIONS="" \
SMQ_SEND_TELEMETRY=true \
SMQ_GROUPS_INSTANCE_ID="" \
$GOBIN/supermq-groups
```

## Usage

Groups supports the following operations:

| Operation                 | Description                                                                 |
| ------------------------- | --------------------------------------------------------------------------- |
| `create`                  | Create a new group within a domain                                          |
| `list`                    | List groups (flat list or tree) with filters for metadata, tags, status     |
| `get`                     | Retrieve a single group (optionally with role memberships)                  |
| `update`                  | Update a group’s name, description, tags, or metadata                       |
| `enable` / `disable`      | Enable or disable a group                                                   |
| `delete`                  | Permanently delete a group                                                  |
| `add-parent` / `remove-parent` | Assign or remove a parent group                                        |
| `add-children` / `remove-children` | Attach or detach child groups (or remove all children)            |
| `list-children`           | List children at specific depth ranges                                      |
| `hierarchy`               | Fetch ancestors/descendants as a tree or list                               |
| `roles`                   | Create/list/update/delete group roles; manage role actions and members      |

### API Examples

#### Create a Group

```bash
curl -X POST http://localhost:9004/<domainID>/groups \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "edge-devices",
    "description": "All edge devices",
    "metadata": { "region": "eu-west-1" },
    "tags": ["iot","edge"],
    "parent_id": "",
    "status": "enabled"
  }'
```

#### List Groups (flat)

```bash
curl -X GET "http://localhost:9004/<domainID>/groups?limit=10&status=enabled" \
  -H "Authorization: Bearer <your_access_token>"
```

#### Retrieve a Group (with Roles)

```bash
curl -X GET "http://localhost:9004/<domainID>/groups/<groupID>?roles=true" \
  -H "Authorization: Bearer <your_access_token>"
```

#### Update a Group

```bash
curl -X PUT http://localhost:9004/<domainID>/groups/<groupID> \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "edge-ops",
    "description": "Edge operations",
    "metadata": { "region": "eu-west-1", "env": "prod" },
    "tags": ["iot","ops"]
  }'
```

#### Enable or Disable a Group

```bash
curl -X POST http://localhost:9004/<domainID>/groups/<groupID>/enable \
  -H "Authorization: Bearer <your_access_token>"

curl -X POST http://localhost:9004/<domainID>/groups/<groupID>/disable \
  -H "Authorization: Bearer <your_access_token>"
```

#### Manage Hierarchy

```bash
# Add a parent
curl -X POST http://localhost:9004/<domainID>/groups/<groupID>/parents \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{ "parent_id": "<parentID>" }'

# List children between levels 1 and 2
curl -X GET "http://localhost:9004/<domainID>/groups/<groupID>/children?start_level=1&end_level=2&limit=10" \
  -H "Authorization: Bearer <your_access_token>"
```

## Roles Management for Groups

Group roles use the shared role manager. Supported operations mirror domain roles (create, list, view, update, delete roles; add/list/remove actions; add/list/remove members; list available actions).

Example: create a group role

```bash
curl -X POST http://localhost:9004/<domainID>/groups/<groupID>/roles \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "group-admin",
    "optional_actions": ["manage_role_permission", "update_permission"],
    "optional_members": ["<userID>"]
  }'
```

List available actions for groups:

```bash
curl -X GET http://localhost:9004/<domainID>/groups/roles/available-actions \
  -H "Authorization: Bearer <your_access_token>"
```

## Implementation Details

- Groups are stored in PostgreSQL with `ltree` paths for hierarchy queries; domain migrations are applied alongside group migrations for referential integrity.
- Role tables are provisioned per entity with a `groups_` prefix.
- Event notifications are published to `SMQ_ES_URL`; domain events are consumed to keep group data aligned.
- Authorization and roles are enforced through SpiceDB and shared policy middleware.
- Optional HTTP callouts (pre-operation hooks) are controlled via `SMQ_GROUPS_CALLOUT_*`.
- Observability: Jaeger tracing, Prometheus metrics at `/metrics`, and a `/health` endpoint.

### Groups Table

| Column        | Type          | Description                                                        |
| ------------- | ------------- | ------------------------------------------------------------------ |
| `id`          | VARCHAR(36)   | UUID of the group (primary key)                                    |
| `parent_id`   | VARCHAR(36)   | Optional parent group (self-referential FK)                        |
| `domain_id`   | VARCHAR(36)   | Owning domain                                                      |
| `name`        | VARCHAR(1024) | Group name                                                         |
| `description` | VARCHAR(1024) | Optional description                                               |
| `metadata`    | JSONB         | Arbitrary metadata                                                 |
| `tags`        | TEXT[]        | Group tags                                                         |
| `path`        | LTREE         | Hierarchical path for fast ancestor/descendant queries             |
| `created_at`  | TIMESTAMPTZ   | Creation timestamp                                                 |
| `updated_at`  | TIMESTAMPTZ   | Last update timestamp                                              |
| `updated_by`  | VARCHAR(254)  | Actor who last updated the group                                   |
| `status`      | SMALLINT      | 0 = enabled, 1 = disabled, 2 = deleted                             |

## Best Practices

- Model hierarchy deliberately: keep depth reasonable and avoid cycles by design.
- Use tags/metadata to segment groups by environment, region, or ownership for filtering.
- Prefer `disable` before `delete` when you need reversible off-boarding.
- Use roles sparingly and audit with `list-role-members`; grant only required actions.
- Fetch children with bounded levels to keep queries efficient.
- Limit callouts to necessary operations via `SMQ_GROUPS_CALLOUT_OPERATIONS`.

## Versioning and Health Check

The Groups service exposes `/health` with status and build metadata.

```bash
curl -X GET http://localhost:9004/health \
  -H "accept: application/health+json"
```

Example response:

```json
{
  "status": "pass",
  "version": "0.18.0",
  "commit": "7d6f4dc4f7f0c1fa3dc24eddfb18bb5073ff4f62",
  "description": "groups service",
  "build_time": "1970-01-01_00:00:00"
}
```

For full API coverage, see the [Groups API documentation](https://docs.api.supermq.absmach.eu/?urls.primaryName=api%2Fgroups.yaml).

[doc]: https://docs.supermq.absmach.eu/
