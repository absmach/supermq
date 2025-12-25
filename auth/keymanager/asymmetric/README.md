# Asymmetric Key Manager

The asymmetric key manager provides EdDSA (Ed25519) signing and verification for SuperMQ authentication tokens. It supports both single-key and multi-key modes with zero-downtime key rotation.

## Features

- **EdDSA (Ed25519) asymmetric signatures** - Secure, fast signing and verification
- **JWKS endpoint support** - Publish public keys for token verification
- **Zero-downtime key rotation** - Overlap keys during rotation for seamless transitions
- **Grace period management** - Configure how long retiring keys remain valid
- **Backward compatible** - Works with existing single-key deployments

## Modes

### Single-Key Mode (Default)

Uses a single private key for signing and verification. This is the default mode when no `keys.json` metadata file exists.

**Directory structure:**
```
/path/to/keys/
└── private.key
```

**Usage:**
```go
km, err := asymmetric.NewKeyManager("/path/to/keys/private.key", idProvider)
```

### Multi-Key Mode (Key Rotation)

Supports multiple keys with different lifecycle states for zero-downtime rotation. Enabled by creating a `keys.json` metadata file.

**Directory structure:**
```
/path/to/keys/
├── keys.json
├── private.key              # Active key
└── private-2024-12-18.key   # Retiring key (during rotation)
```

**Usage:**
```go
km, err := asymmetric.NewKeyManager("/path/to/keys/private.key", idProvider)
// Automatically detects keys.json and switches to multi-key mode
```

## Key Lifecycle States

### Active
- Used for **signing** new tokens
- Used for **verification** of existing tokens
- Only one key can be active at a time
- Must be referenced by `active_key_id` in metadata
- **MAKE SURE ACTIVE KEY IS NOT EXPIRED**

### Retiring
- **NOT** used for signing new tokens
- Still used for **verification** of tokens signed before rotation
- Valid until `expires_at` timestamp (grace period)
- Allows seamless rotation without invalidating existing tokens

### Retired
- No longer used for signing or verification
- Automatically filtered from JWKS endpoint
- Can be safely deleted from disk

## Key Rotation Guide

### 1. Initial Setup (Single Key)

Start with a single active key:

```bash
# Generate Ed25519 key pair
openssl genpkey -algorithm Ed25519 -out private.key
```

**Directory:**
```
/keys/
└── private.key
```

### 2. Prepare for Rotation

Generate a new key and create `keys.json` metadata:

```bash
# Generate new key
openssl genpkey -algorithm ED25519 -out private-new.key
```

**Create `keys.json`:**
```json
{
  "active_key_id": "key-2024-12-25",
  "grace_period_hours": 168,
  "keys": [
    {
      "id": "key-2024-12-25",
      "file": "private-new.key",
      "created_at": "2024-12-25T00:00:00Z",
      "status": "active"
    },
    {
      "id": "key-2024-12-18",
      "file": "private.key",
      "created_at": "2024-12-18T00:00:00Z",
      "status": "retiring",
      "expires_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

**Directory:**
```
/keys/
├── keys.json
├── private-new.key    # Active (new)
└── private.key        # Retiring (old)
```

### 3. Deploy and Restart

Restart the auth service. The key manager will:
- Load both keys from metadata
- Sign new tokens with `active` (active)
- Verify tokens with both keys (active + retiring)
- Publish both keys in JWKS endpoint

### 4. Grace Period

During the grace period (default 168 hours / 7 days):
- New tokens are signed with the active key
- Old tokens remain valid (verified with retiring key)
- JWKS endpoint returns both public keys
- Clients gradually refresh tokens with new key

### 5. Complete Rotation

After the grace period expires:

```json
{
  "active_key_id": "key-2024-12-25",
  "grace_period_hours": 168,
  "keys": [
    {
      "id": "key-2024-12-25",
      "file": "private-new.key",
      "created_at": "2024-12-25T00:00:00Z",
      "status": "active"
    }
  ]
}
```

**Cleanup:**
```bash
# Remove expired key
rm private.key
```

**Directory:**
```
/keys/
├── keys.json
└── private-new.key
```

Restart the service. Old key is no longer loaded.

## Metadata Reference

### keys.json Schema

```json
{
  "active_key_id": "string (required)",
  "grace_period_hours": "int (optional, default: 168)",
  "keys": [
    {
      "id": "string (required, unique)",
      "file": "string (required, relative to keys directory)",
      "created_at": "RFC3339 timestamp (required)",
      "status": "active|retiring|expired (required)",
      "expires_at": "RFC3339 timestamp (required for retiring keys)"
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `active_key_id` | string | Yes | ID of the key used for signing. Must reference a key with `"status": "active"` |
| `grace_period_hours` | int | No | Hours a retiring key remains valid. Default: 168 (7 days) |
| `keys[].id` | string | Yes | Unique identifier for the key. Used as JWT `kid` header |
| `keys[].file` | string | Yes | Filename relative to keys directory |
| `keys[].created_at` | RFC3339 | Yes | When the key was created |
| `keys[].status` | enum | Yes | Lifecycle state: `active`, `retiring`, or `expired` |
| `keys[].expires_at` | RFC3339 | Yes* | When the key expires. *Required for `retiring` status |

### Validation Rules

- Only one key can have `"status": "active"`
- `active_key_id` must reference an existing key with active status
- Retiring keys must have `expires_at` set
- Key IDs must be unique
- All referenced key files must exist

## Best Practices

### Grace Period Configuration

**Recommended:** 168 hours (7 days)
- Allows time for all clients to refresh tokens
- Covers weekend deployments
- Provides buffer for unforeseen issues

**Minimum:** 24 hours
- Only for controlled environments
- Requires careful monitoring of token refresh rates

**Maximum:** 720 hours (30 days)
- For critical production systems
- When token lifetime is very long
- Extra safety margin for rotation

### Key Rotation Schedule

**Recommended frequency:**
- **Every 90 days** for standard deployments
- **Every 30 days** for high-security environments
- **Immediately** if key compromise is suspected

### Key Naming Convention

Use date-based naming for clarity:
```
private-YYYY-MM-DD.key
```

Example:
```
private-2024-12-25.key
private-2024-12-18.key
```

### Security Considerations

1. **File Permissions:** Ensure private keys are readable only by the auth service
   ```bash
   chmod 600 *.key
   chown auth-service:auth-service *.key
   ```

2. **Backup Strategy:** Securely backup retiring keys until grace period expires

3. **Key Generation:** Use cryptographically secure random number generators
   ```bash
   openssl genpkey -algorithm ED25519 -out private.key
   ```

4. **Secrets Management:** Consider using a secrets manager for production deployments

5. **Audit Logging:** Monitor key rotation events and failed verifications

## Migration from Single-Key to Multi-Key

### Step 1: Backup Current Key

```bash
cp /path/to/keys/private.key /path/to/keys/private-backup.key
```

### Step 2: Create Metadata

Create `keys.json` referencing the existing key:

```json
{
  "active_key_id": "key-initial",
  "grace_period_hours": 168,
  "keys": [
    {
      "id": "key-initial",
      "file": "private.key",
      "created_at": "2024-12-18T00:00:00Z",
      "status": "active"
    }
  ]
}
```

### Step 3: Restart and Verify

```bash
# Restart auth service
systemctl restart supermq-auth

# Verify JWKS endpoint returns the key
curl http://localhost:8189/.well-known/jwks.json
```

### Step 4: Ready for Rotation

You can now follow the standard rotation process to add new keys.

## Troubleshooting

### "active key not loaded successfully"

**Cause:** The key referenced by `active_key_id` couldn't be loaded.

**Solutions:**
- Verify the key file exists and has correct permissions
- Check that `active_key_id` matches a key `id` in the `keys` array
- Ensure the referenced key has `"status": "active"`

### "failed to load private key"

**Cause:** Key file is missing or unreadable.

**Solutions:**
- Verify file path is correct (relative to keys directory)
- Check file permissions (should be readable by auth service)
- Ensure key is in valid PEM or raw Ed25519 format

### "invalid key status"

**Cause:** Unknown status value in metadata.

**Solutions:**
- Use only: `active`, `retiring`, or `retired`
- Check for typos in status field

### Tokens signed with old key fail verification

**Cause:** Retiring key expired or not loaded.

**Solutions:**
- Check `expires_at` timestamp hasn't passed
- Verify key file exists and is loaded
- Check logs for key loading errors

## Example Deployment

### Docker Compose

```yaml
services:
  auth:
    image: supermq/auth:latest
    volumes:
      - ./keys:/keys:ro
    environment:
      SMQ_AUTH_PRIVATE_KEY_PATH: /keys/private.key
```

**Directory structure:**
```
./keys/
├── keys.json
├── private.key
└── private-2024-12-18.key
```

### Kubernetes Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-keys
type: Opaque
data:
  keys.json: <base64-encoded-metadata>
  private.key: <base64-encoded-active-key>
  private-2024-12-18.key: <base64-encoded-retiring-key>
---
apiVersion: v1
kind: Pod
metadata:
  name: auth
spec:
  containers:
  - name: auth
    image: supermq/auth:latest
    volumeMounts:
    - name: keys
      mountPath: /keys
      readOnly: true
    env:
    - name: SMQ_AUTH_PRIVATE_KEY_PATH
      value: /keys/private.key
  volumes:
  - name: keys
    secret:
      secretName: auth-keys
      defaultMode: 0400
```

## API Reference

### KeyManager Interface

```go
type KeyManager interface {
    // Sign creates a JWT token from the provided key using the active key
    Sign(key auth.Key) (string, error)

    // Verify validates a JWT token using any valid key (active or retiring)
    Verify(tokenString string) (auth.Key, error)

    // PublicKeys returns all valid public keys for the JWKS endpoint
    PublicKeys() ([]auth.PublicKeyInfo, error)
}
```

### Creating a Key Manager

```go
import (
    "github.com/absmach/supermq/auth/keymanager/asymmetric"
)

// Single-key mode (no keys.json)
km, err := asymmetric.NewKeyManager("/path/to/private.key", idProvider)

// Multi-key mode (with keys.json in same directory)
km, err := asymmetric.NewKeyManager("/path/to/keys/private.key", idProvider)
```

### Signing Tokens

```go
key := auth.Key{
    ID:        "token-123",
    Type:      auth.AccessKey,
    Subject:   "user-456",
    Role:      auth.UserRole,
    IssuedAt:  time.Now().UTC(),
    ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
    Verified:  true,
}

token, err := km.Sign(key)
if err != nil {
    // Handle error
}
```

### Verifying Tokens

```go
key, err := km.Verify(token)
if err != nil {
    // Token is invalid or expired
}

// Use key.Subject, key.Role, etc.
```

### Publishing JWKS

```go
publicKeys, err := km.PublicKeys()
if err != nil {
    // Handle error
}

// Serialize to JSON for JWKS endpoint
jwks := map[string]interface{}{
    "keys": publicKeys,
}
```
