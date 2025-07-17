# SuperMQ Certs with OpenBao PKI

This addon configures SuperMQ certificate service to use OpenBao as the PKI backend instead of AM Certs or Vault.

## Quick Setup

1. Start OpenBao first:
   ```bash
   docker compose -f docker/docker-compose.yaml -f docker/addons/certs/docker-compose.yaml up openbao -d
   ```

2. Wait for OpenBao to be ready, then run the initialization script:
   ```bash
   docker exec -it supermq-openbao /bin/sh /docker/addons/certs/init-openbao.sh
   ```

3. The script will output the AppRole credentials. Update your `.env` file with the provided values:
   ```bash
   SMQ_CERTS_OPENBAO_APP_ROLE=<role-id-from-script>
   SMQ_CERTS_OPENBAO_APP_SECRET=<secret-id-from-script>
   ```

4. Start the certs service:
   ```bash
   docker compose -f docker/docker-compose.yaml -f docker/addons/certs/docker-compose.yaml up certs -d
   ```

## Manual Setup (Alternative)

If you prefer to set up OpenBao manually:

1. Access OpenBao UI at http://localhost:8200
2. Login with root token: `openbao-root-token` (or your custom token)
3. Enable AppRole authentication:
   ```bash
   bao auth enable approle
   ```
4. Enable PKI secrets engine:
   ```bash
   bao secrets enable -path=pki pki
   ```
5. Configure PKI and create the SuperMQ role and policy as shown in the init script.

## Configuration

The following environment variables are used to configure OpenBao PKI:

- `SMQ_CERTS_OPENBAO_HOST`: OpenBao server URL (default: `http://supermq-openbao:8200`)
- `SMQ_CERTS_OPENBAO_APP_ROLE`: AppRole role ID for authentication
- `SMQ_CERTS_OPENBAO_APP_SECRET`: AppRole secret ID for authentication  
- `SMQ_CERTS_OPENBAO_NAMESPACE`: OpenBao namespace (optional)
- `SMQ_CERTS_OPENBAO_PKI_PATH`: PKI secrets engine path (default: `pki`)
- `SMQ_CERTS_OPENBAO_ROLE`: PKI role name for certificate issuance (default: `supermq`)
- `SMQ_OPENBAO_ROOT_TOKEN`: Root token for development mode (default: `openbao-root-token`)

## Development Mode

The OpenBao container runs in development mode with:
- In-memory storage
- TLS disabled  
- UI enabled at http://localhost:8200
- Root token authentication

**Warning**: Development mode should never be used in production as data is not persisted and security is disabled.

## Production Setup

For production, you should:
1. Use a proper storage backend (not in-memory)
2. Enable TLS
3. Configure proper authentication methods
4. Set up proper PKI hierarchy
5. Use secure tokens and secrets

Refer to the [OpenBao documentation](https://openbao.org/docs/) for production deployment guidance.
