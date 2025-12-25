# Asymmetric Key Manager

EdDSA (Ed25519) key manager with support for zero-downtime key rotation.

## Features

- **Single-key mode** - Simple setup with one private key (default)
- **Multi-key mode** - Multiple keys with overlapping validity for rotation
- **Zero-downtime rotation** - Active + retiring keys work simultaneously during grace period
- **JWKS endpoint** - Publishes all valid public keys for token verification

Key rotation is offloaded to the external service. **Key expiration is checked only on service startup**.
If you want to invalidate or rotate keys, a manual update to keys directory and `keys.json` file and 
by service restart are required. 

## How It Works

### Single-Key Mode

Without a `keys.json` file, the manager operates in single-key mode using only the private key file specified in the configuration directory.

### Multi-Key Mode

With a `keys.json` file in the same directory as the private key, the manager enables key rotation:

1. **Active key** - Used for signing new tokens and verification
2. **Retiring keys** - Used only for verification during grace period
3. **Retired keys** - Expired retiring keys, filtered from JWKS

### Key Lifecycle

**Active** → Sign new tokens + verify existing tokens
**Retiring** → Verify only (grace period active)
**Retired** → No longer used (grace period expired)

## Configuration

### Metadata File: `keys.json`

```json
{
  "active_key_id": "key-2025-12-25",
  "keys": [
    {
      "id": "key-2025-12-25",
      "file": "private-new.key",
      "created_at": "2024-12-25T00:00:00Z",
      "status": "active"
    },
    {
      "id": "key-2024-12-25",
      "file": "private.key",
      "created_at": "2024-12-25T00:00:00Z",
      "status": "retiring",
      "expires_at": "2026-06-01T00:00:00Z"
    }
  ]
}
```

### Field Reference

| Field               | Required     | Description                               |
| ------------------- | ------------ | ----------------------------------------- |
| `active_key_id`     | Yes          | ID of the active key used for signing     |
| `keys[].id`         | Yes          | Unique key identifier (used as JWT `kid`) |
| `keys[].file`       | Yes          | Key filename relative to keys directory   |
| `keys[].created_at` | Yes          | Key creation timestamp (RFC3339)          |
| `keys[].status`     | Yes          | `active`, `retiring`, or `retired`        |
| `keys[].expires_at` | For retiring | When retiring key expires (RFC3339)       |

### Validation Rules

- Only one key can be active
- Active key must not be expired
- Retiring keys must have `expires_at` set
- Active keys should not have `expires_at` set
- All key files must exist and be readable

## Key Rotation Process

1. **Generate new key** - Create new Ed25519 private key
2. **Update metadata** - Set old key to `retiring` with `expires_at = now + grace_period`, set new key as `active`
3. **Restart service** - Both keys are loaded, new key signs, both verify
4. **Wait for grace period** - Old tokens remain valid
5. **Clean up** - Remove expired retiring key from metadata and delete file

### Grace Period Recommendation

When setting `expires_at` for retiring keys, calculate as:

```
expires_at = rotation_time + grace_period
```

**Recommended grace period:** 168 hours (7 days)
**Minimum:** 24 hours
**Maximum:** 720 hours (30 days)

## Security

- Store private keys with `0600` permissions
- Use cryptographically secure key generation (`openssl genpkey -algorithm Ed25519`)
- Rotate keys every 90 days (30 days for high-security environments)
- Never commit keys to version control - the existing keys are only for the demonstration of the JWKS feature in the platform
- Consider using secrets management in production

## Migration

To enable rotation on an existing single-key deployment, create `keys.json` with your current key marked as `active`. The manager automatically switches to multi-key mode when the metadata file is present.
