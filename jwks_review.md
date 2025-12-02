# JWKS/Asymmetric Key Authentication Branch Code Review

## Overview
This branch (smq1672-token) replaces HMAC-based symmetric JWT authentication with RSA asymmetric key authentication and adds a JWKS (JSON Web Key Set) endpoint. This is a **major security architecture change** that enables proper distributed authentication and token verification without sharing secrets across services.

**Branch**: `smq1672-token`
**Base**: `main`
**Files Changed**: 30 files (+1896, -542)

---

## Architecture Changes

### From Symmetric to Asymmetric Cryptography

**Before (HMAC/HS512)**:
- Single shared secret (`SMQ_AUTH_SECRET_KEY`)
- All services need the same secret to verify tokens
- Secret distribution problem
- Token signing and verification use same key

**After (RSA/RS256)**:
- Private/public key pairs
- Only auth service has private keys
- Other services fetch public keys via JWKS endpoint
- Better security isolation

### New Components

1. **KeyManager** (`auth/keymanager.go`, `auth/keymanager/keymanager.go`)
   - Manages RSA key pairs lifecycle
   - Handles key rotation
   - Supports active, next, and retired keys
   - Optional file-based persistence

2. **JWKS Endpoint** (`auth/api/http/keys/`)
   - Exposes public keys at `/.well-known/jwks.json`
   - Standard OAuth2/OIDC discovery endpoint
   - Cache headers for performance

3. **JWKS Authentication** (`pkg/authn/jwks/authn.go`)
   - Client-side authentication using JWKS
   - Fetches and caches public keys
   - Validates tokens without secrets

---

## Key Changes

### 1. KeyManager Implementation

**auth/keymanager/keymanager.go** (363 lines, new)

Key features:
- **Three-key rotation model**:
  - `activeID`: Current signing key
  - `nextID`: Pre-generated next key
  - `retiredID`: Previous key (within grace period)

- **Automatic rotation**:
  - Configurable interval (default 24h)
  - Background goroutine handles rotation
  - Grace period = login duration + 15 seconds

- **File persistence**:
  - Optional saving to `keys/keys.json`
  - Atomic writes with temp file + rename
  - Loads existing keys on startup

- **Thread safety**:
  - RWMutex for concurrent access
  - Read lock for signing/parsing
  - Write lock for rotation

**Key Methods**:
```go
SignJWT(token jwt.Token) ([]byte, error)      // Sign with active private key
ParseJWT(token string) (jwt.Token, error)      // Verify with active + retired keys
PublicJWKS() []jwk.Key                         // Return public keys for JWKS
Rotate() error                                  // Perform key rotation
```

### 2. Tokenizer Refactoring

**auth/jwt/tokenizer.go**

Changes:
- Removed `secret []byte` field
- Added `keyManager auth.KeyManager` dependency
- `SignJWT` now delegates to KeyManager
- `ParseJWT` uses KeyManager (supports key rotation)
- New `RetrieveJWKS()` method
- Made `ToKey()` exported for reuse

**auth/tokenizer.go** (interface)
- Added `RetrieveJWKS() []jwk.Key` method

### 3. JWKS Endpoint

**auth/api/http/keys/transport.go:49-55**
```go
r.Get("/.well-known/jwks.json", kithttp.NewServer(
    retrieveJWKSEndpoint(svc),
    decodeKeyReq,
    api.EncodeResponse,
    opts...,
).ServeHTTP)
```

**auth/api/http/keys/responses.go:74-92**
- Response includes `keys` array of JWK public keys
- Cache headers: `public, max-age=900, stale-while-revalidate=60` (15 min cache)

### 4. JWKS Client Authentication

**pkg/authn/jwks/authn.go** (133 lines, new)

Features:
- Fetches JWKS from configurable URL
- 5-minute in-memory cache
- Validates issuer (`supermq.auth`)
- Validates token expiry
- Extracts session from token

**Usage in services** (clients, channels, domains, groups, http, journal, users, ws):
```go
// Before (authsvc gRPC client)
authn, authnClient, err := authsvcAuthn.NewAuthentication(ctx, grpcCfg)
defer authnClient.Close()

// After (JWKS)
authn := jwksAuthn.NewAuthentication(jwksURL)
```

Benefits:
- No gRPC connection needed
- Simpler configuration
- Lower latency (local verification after cache)
- Independent service deployment

### 5. Configuration Changes

**docker/.env**:
- ❌ Removed: `SMQ_AUTH_SECRET_KEY`
- ✅ Added: `SMQ_AUTH_KEYS_ROTATION_INTERVAL` (default: "24h")
- ✅ Added: `SMQ_AUTH_KEYS_SAVE_TO_FILE` (default: "true")

**docker/docker-compose.yaml**:
- Added volume: `supermq-auth-keys-volume` mounted at `/keys`
- Removed `SMQ_AUTH_SECRET_KEY` environment variable
- Added key rotation config variables

### 6. Service Changes

**cmd/auth/main.go**:
- Creates `KeyManager` with ULID ID provider
- Passes `KeyManager` to tokenizer
- Removed secret key from config

**All other services** (clients, channels, domains, groups, http, journal, users, ws):
- Replaced `authsvcAuthn.NewAuthentication()` with `jwksAuthn.NewAuthentication()`
- Hardcoded JWKS URL: `http://auth:9001/keys/.well-known/jwks.json`
- Removed gRPC client connection for authentication
- Still use gRPC for authorization

### 7. Test Updates

**auth/jwt/token_test.go**:
- Updated to use mocked `KeyManager`
- Added `signToken()` helper function
- Tests key signing and parsing with rotation scenarios

**auth/service_test.go**:
- Refactored to use mocked `Tokenizer`
- Updated test setup to create keys properly
- Tests cover new authentication flow

---

## Security Analysis

### ✅ Security Improvements

1. **Proper Secret Separation**
   - Private keys never leave auth service
   - No shared secrets across services
   - Follows OAuth2/OIDC best practices

2. **Key Rotation**
   - Automatic periodic rotation
   - Graceful handling of old tokens
   - Reduces impact of key compromise

3. **Standard Compliance**
   - JWKS endpoint follows RFC 7517
   - Standard `.well-known` discovery path
   - Compatible with standard JWT libraries

4. **Defense in Depth**
   - Even if public keys leaked, can't forge tokens
   - Key rotation limits exposure window
   - File permissions (0o600) protect private keys

### ⚠️ Security Concerns

#### 1. **CRITICAL: Unencrypted JWKS Endpoint**

**Location**: All services hardcode `http://auth:9001/keys/.well-known/jwks.json`

**Issue**: Using HTTP instead of HTTPS for JWKS endpoint opens MITM attack vector:
- Attacker can intercept and replace public keys
- Could sign malicious tokens with their own keys
- Services would accept forged tokens

**Impact**: HIGH - Completely bypasses authentication

**Mitigation**:
- Use HTTPS for JWKS URL in production
- Or ensure network isolation (service mesh, VPN)
- Add TLS certificate pinning
- Consider serving JWKS on same port as main API with TLS

#### 2. **Race Condition in JWKS Cache**

**Location**: `pkg/authn/jwks/authn.go:37-40`

```go
var (
    jwksCache = struct {
        jwks     jwk.Set
        cachedAt time.Time
    }{}  // ❌ No mutex protection
)
```

**Issue**:
- Global variable accessed from multiple goroutines
- `fetchJWKS()` reads and writes without locking (line 85-104)
- Concurrent requests could cause data races

**Impact**: MEDIUM - Potential panics, incorrect cache behavior

**Fix**:
```go
var jwksCache = struct {
    sync.RWMutex
    jwks     jwk.Set
    cachedAt time.Time
}{}
```

#### 3. **No JWKS Signature Verification**

**Location**: `pkg/authn/jwks/authn.go:78-106`

The JWKS endpoint response is not authenticated:
- No signature verification
- No TLS certificate validation shown
- Trust on first use (TOFU) problem

**Impact**: MEDIUM - Susceptible to MITM

**Recommendation**:
- Implement JWKS signature verification (JWS)
- Or require mutual TLS
- Or pre-configure public key fingerprints

#### 4. **Weak Random Number Generator for Keys**

**Location**: `auth/keymanager/keymanager.go:329`

```go
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
```

✅ **GOOD**: Using `crypto/rand` (cryptographically secure)

However:
- No entropy check before generation
- No FIPS compliance mentioned
- 2048-bit keys (acceptable but 4096 preferred for long-term)

**Recommendation**: Document key strength requirements and rotation policy

#### 5. **Key Material in Memory**

**Issue**: Private keys stored unencrypted in memory
- Vulnerable to memory dumps
- No memory locking (mlock)
- Go's GC may copy keys around memory

**Impact**: LOW - Requires host compromise

**Recommendation** (for high-security environments):
- Use HSM or KMS for key storage
- Implement memory zeroing on key rotation
- Consider using secure enclaves

#### 6. **File Permissions Race**

**Location**: `auth/keymanager/keymanager.go:322-325`

```go
tmp := km.statePath + ".tmp"
if err := os.WriteFile(tmp, data, 0o600); err != nil {
    return err
}
return os.Rename(tmp, km.statePath)  // ❌ Final file inherits umask
```

**Issue**:
- `os.Rename()` doesn't preserve file permissions
- Final file may have world-readable permissions if umask is permissive

**Fix**:
```go
if err := os.WriteFile(tmp, data, 0o600); err != nil {
    return err
}
if err := os.Chmod(tmp, 0o600); err != nil {
    return err
}
return os.Rename(tmp, km.statePath)
```

Or better:
```go
if err := os.WriteFile(km.statePath, data, 0o600); err != nil {
    return err
}
// Atomic writes with sync
f.Sync()
```

---

## Code Quality Issues

### 1. **Error Swallowed in Rotation Handler**

**Location**: `auth/keymanager/keymanager.go:237-250`

```go
func (km *manager) rotateHandler(ctx context.Context) error {
    ticker := time.NewTicker(km.rotationInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            if err := km.Rotate(); err != nil {
                return err  // ❌ Goroutine exits, rotation stops forever
            }
        }
    }
}
```

**Issue**: If rotation fails, the goroutine exits and automatic rotation stops permanently

**Impact**: HIGH - System will fail after current key expires

**Fix**:
```go
case <-ticker.C:
    if err := km.Rotate(); err != nil {
        // Log error but continue trying
        logger.Error("key rotation failed", slog.Any("error", err))
        // Optional: exponential backoff, alerts, etc.
    }
```

### 2. **Ignored Error in PublicJWKS**

**Location**: `auth/keymanager/keymanager.go:186`

```go
if km.saveToFile {
    _ = km.saveToDisk()  // ❌ Error ignored
}
```

**Issue**: Disk write failure during cleanup is silently ignored

**Impact**: LOW - State may be inconsistent after restart

**Recommendation**: Log the error at minimum

### 3. **Typo in Variable Name**

**Location**: `auth/keymanager/keymanager.go:30`

```go
var errEmptyKerDir = errors.New("key directory cannot be empty when saving to file")
//              ^^^
//              Should be: errEmptyKeyDir
```

**Impact**: None (cosmetic)

### 4. **Magic Numbers**

**Location**: `auth/keymanager/keymanager.go:26, 329`

```go
const skewDuration = 15 * time.Second  // Why 15?

privateKey, err := rsa.GenerateKey(rand.Reader, 2048)  // Why 2048?
```

**Recommendation**: Add comments explaining rationale or make configurable

### 5. **Hardcoded JWKS URL**

**Location**: All service main.go files

```go
const jwksURL = "http://auth:9001/keys/.well-known/jwks.json"
```

**Issues**:
- Not configurable via environment
- Hardcoded HTTP scheme
- Hardcoded port
- No fallback if auth service moves

**Impact**: MEDIUM - Deployment inflexibility

**Fix**: Make configurable:
```go
type config struct {
    JWKSUrl string `env:"SMQ_AUTH_JWKS_URL" envDefault:"https://auth:9001/keys/.well-known/jwks.json"`
}
```

### 6. **Inefficient Lock Usage**

**Location**: `auth/keymanager/keymanager.go:174-196`

```go
func (km *manager) PublicJWKS() []jwk.Key {
    km.mu.Lock()  // ❌ Write lock when read lock sufficient
    defer km.mu.Unlock()

    keys := []jwk.Key{km.keySet[km.activeID].publicKey}
    // ... only reads until line 181 ...

    if time.Since(kp.retiredAt) > km.gracePeriod {
        delete(km.keySet, km.retiredID)  // ✅ Write happens here
```

**Issue**: Uses write lock for entire function when most is read-only

**Fix**: Use read lock initially, upgrade to write lock only when deleting:
```go
km.mu.RLock()
keys := []jwk.Key{km.keySet[km.activeID].publicKey}
// ... read operations ...
shouldCleanup := km.retiredID != "" && time.Since(km.keySet[km.retiredID].retiredAt) > km.gracePeriod
km.mu.RUnlock()

if shouldCleanup {
    km.mu.Lock()
    // Re-check condition after acquiring write lock
    delete(km.keySet, km.retiredID)
    km.retiredID = ""
    km.mu.Unlock()
}
```

### 7. **Duplicate Code in Services**

All service main.go files have identical JWKS setup:
```go
authn := jwksAuthn.NewAuthentication(jwksURL)
logger.Info("AuthN successfully set up jwks authentication on " + jwksURL)
```

**Recommendation**: Extract to shared initialization function

---

## Performance Considerations

### 1. **RSA Performance**

**Issue**: RSA operations are slower than HMAC
- RSA signing: ~1000x slower than HMAC
- RSA verification: ~100x slower than HMAC

**Impact**:
- Higher CPU usage on auth service (signing)
- Lower impact on other services (verification still fast with caching)

**Mitigation**:
- JWKS cache reduces verification overhead
- Consider shorter token expiry to reduce signing load
- Monitor auth service CPU usage

### 2. **Key Rotation Performance**

**Location**: `auth/keymanager/keymanager.go:199-235`

**Issue**: Rotation generates new 2048-bit RSA key (CPU intensive)
- Takes ~50-100ms on modern CPU
- Blocks with write lock

**Impact**: LOW (happens every 24h by default)

**Optimization**: Generate next key in background before rotation:
```go
// Pre-generate new key without lock
newPair, err := generateKeyPair(newID)
if err != nil {
    return err
}

// Quick swap with lock
km.mu.Lock()
km.keySet[newID] = newPair
// ... rest of rotation ...
km.mu.Unlock()
```

### 3. **JWKS Cache Strategy**

**Location**: `pkg/authn/jwks/authn.go:85-86`

```go
if time.Since(jwksCache.cachedAt) < cacheDuration && jwksCache.jwks.Len() > 0 {
    return jwksCache.jwks, nil
}
```

**Issues**:
- All requests see cache miss at same time after expiry (thundering herd)
- No background refresh
- 5-minute cache seems short for keys that rotate every 24h

**Recommendations**:
- Increase cache duration to 30-60 minutes
- Implement stale-while-revalidate pattern
- Use HTTP cache headers from response
- Add cache metrics

### 4. **File I/O on Every Rotation**

**Location**: `auth/keymanager/keymanager.go:228-231`

```go
if km.saveToFile {
    if err := km.saveToDisk(); err != nil {
        return err
    }
}
```

**Issue**: Synchronous file write during rotation (holds write lock)

**Optimization**: Save to disk asynchronously after rotation:
```go
if km.saveToFile {
    go func() {
        if err := km.saveToDisk(); err != nil {
            // Log error
        }
    }()
}
```

---

## Testing Assessment

### ✅ Strengths

1. **Comprehensive Test Updates**
   - Token tests updated for KeyManager mocking
   - Service tests refactored for new architecture
   - Mock implementations generated

2. **Good Mock Coverage**
   - `auth/mocks/key_manager.go` (256 lines)
   - `auth/mocks/tokenizer.go` (208 lines)
   - All interfaces properly mocked

### ❌ Gaps

1. **No KeyManager Unit Tests**
   - 363-line implementation with no tests
   - Key rotation logic untested
   - File persistence untested
   - Concurrency untested

2. **No JWKS Client Tests**
   - `pkg/authn/jwks/authn.go` has no tests
   - Cache behavior untested
   - Error handling untested
   - Issuer validation untested

3. **No Integration Tests**
   - End-to-end token flow untested
   - Key rotation with active tokens untested
   - JWKS endpoint + client integration untested

4. **No Key Rotation Tests**
   - Grace period behavior untested
   - Retired key cleanup untested
   - Concurrent rotation untested

5. **No Security Tests**
   - Token forgery attempts untested
   - Expired token handling untested
   - Invalid signature handling untested

### Recommended Tests

#### KeyManager Tests
```go
func TestKeyManagerRotation(t *testing.T)
func TestKeyManagerGracePeriod(t *testing.T)
func TestKeyManagerConcurrency(t *testing.T)
func TestKeyManagerFilePersistence(t *testing.T)
func TestKeyManagerLoadFromDisk(t *testing.T)
```

#### JWKS Client Tests
```go
func TestJWKSAuthentication(t *testing.T)
func TestJWKSCache(t *testing.T)
func TestJWKSInvalidIssuer(t *testing.T)
func TestJWKSExpiredToken(t *testing.T)
func TestJWKSFetchError(t *testing.T)
```

#### Integration Tests
```go
func TestEndToEndTokenFlow(t *testing.T)
func TestKeyRotationWithActiveTokens(t *testing.T)
func TestMultiServiceAuthentication(t *testing.T)
```

---

## Migration & Deployment

### ⚠️ Breaking Changes

1. **Configuration**
   - `SMQ_AUTH_SECRET_KEY` removed (will break existing deployments)
   - New keys configuration required
   - All services need updated configuration

2. **Token Format**
   - Tokens signed with RSA instead of HMAC
   - Old tokens will be invalid after deployment
   - All users need to re-authenticate

3. **Service Dependencies**
   - Services now depend on JWKS endpoint availability
   - Network requirements changed (HTTP instead of gRPC for authn)

### Migration Strategy

#### Option 1: Blue-Green Deployment (Recommended)
1. Deploy new auth service with key generation
2. Configure JWKS URL in new service instances
3. Switch traffic to new instances
4. Deprecate old instances after token expiry

**Downtime**: ~0 (users need to re-login)

#### Option 2: Gradual Migration
1. Support both HMAC and RSA in auth service temporarily
2. Update services one by one to use JWKS
3. Remove HMAC support after all services updated

**Complexity**: HIGH (requires dual-mode authentication)

### Deployment Checklist

- [ ] Backup existing secret key (rollback plan)
- [ ] Configure key rotation interval
- [ ] Set up persistent volume for keys
- [ ] Update all service configurations
- [ ] Update JWKS URL to use HTTPS
- [ ] Configure TLS for auth service
- [ ] Test key rotation in staging
- [ ] Monitor auth service CPU usage
- [ ] Set up alerts for key rotation failures
- [ ] Document key recovery procedure
- [ ] Plan user re-authentication communication

### Rollback Plan

If issues arise:
1. Revert to previous version with HMAC
2. Restore `SMQ_AUTH_SECRET_KEY` configuration
3. Users will need to re-authenticate again

**Critical**: Key rotation makes rollback difficult after 24h

---

## Operational Concerns

### 1. **Key Recovery**

**Scenario**: Key file corrupted/lost

**Current Behavior**: System generates new keys, all tokens invalid

**Impact**: All users forced to re-authenticate

**Recommendations**:
- Implement key backup strategy
- Store keys in secrets manager (Vault, AWS Secrets Manager)
- Document recovery procedure
- Monitor key file integrity

### 2. **Key Rotation Monitoring**

**Missing**:
- No metrics for rotation success/failure
- No alerts on rotation failures
- No monitoring of key age

**Recommendations**:
```go
// Add metrics
rotations_total{status="success|failure"}
active_key_age_seconds
retired_key_grace_period_remaining_seconds
```

### 3. **JWKS Endpoint Availability**

**Issue**: If JWKS endpoint is down, all authentication fails

**Impact**: CRITICAL - System-wide authentication outage

**Mitigations**:
- Health check for JWKS endpoint
- Monitor JWKS fetch errors
- Implement circuit breaker in JWKS client
- Consider local public key caching with longer TTL

### 4. **Disk Space for Keys**

**Issue**: Key files grow over time if retired keys not cleaned up

**Current**: Retired keys deleted after grace period ✅

**Monitoring**: Watch `/keys` volume usage

### 5. **Clock Skew**

**Issue**: Grace period assumes synchronized clocks

**Current**: 15-second skew allowance ✅

**Recommendation**: Document NTP requirement

---

## Documentation Needs

### Missing Documentation

1. **Architecture Decision Record (ADR)**
   - Why move to asymmetric keys?
   - Trade-offs considered?
   - Performance impact assessment?

2. **Key Management Guide**
   - How to back up keys?
   - Key rotation schedule?
   - Recovery procedures?

3. **Migration Guide**
   - Step-by-step migration from HMAC
   - Downtime expectations
   - Rollback procedure

4. **Operational Runbook**
   - How to manually rotate keys?
   - How to recover from key loss?
   - Troubleshooting JWKS errors

5. **Security Documentation**
   - Threat model
   - Key strength rationale (2048-bit)
   - Compliance considerations (FIPS, etc.)

### Recommended Documentation

#### docs/authentication.md
```markdown
# Authentication Architecture

## Overview
SuperMQ uses asymmetric RSA-2048 JWT authentication...

## Key Management
- Keys rotate every 24 hours
- Grace period: login duration + 15s
- Retired keys accepted during grace period

## JWKS Endpoint
- Location: https://auth:9001/keys/.well-known/jwks.json
- Cache: 15 minutes
- Format: RFC 7517

## Operations
- Manual rotation: ...
- Backup procedure: ...
- Recovery: ...
```

---

## Recommendations

### 🔴 Critical (Must Fix Before Merge)

1. **Fix JWKS cache race condition** (add mutex)
2. **Fix rotation handler error handling** (log and continue, don't exit)
3. **Add KeyManager unit tests** (minimum: rotation, parsing, grace period)
4. **Make JWKS URL configurable** (not hardcoded)
5. **Document migration plan** (breaking change requires clear communication)

### 🟠 High Priority

6. **Add JWKS client tests** (authentication, caching, error cases)
7. **Fix file permissions race** in saveToDisk
8. **Use HTTPS for JWKS URLs** (or document network security requirements)
9. **Add monitoring and metrics** for key rotation
10. **Document key management** (backup, recovery, rotation)

### 🟡 Medium Priority

11. **Optimize lock usage** in PublicJWKS (read lock optimization)
12. **Increase JWKS cache duration** (30-60 minutes)
13. **Add integration tests** (end-to-end token flow)
14. **Implement background key pre-generation** (performance)
15. **Add circuit breaker** to JWKS client

### 🟢 Low Priority (Nice to Have)

16. Consider 4096-bit RSA keys for long-term security
17. Add JWKS signature verification
18. Implement stale-while-revalidate for JWKS cache
19. Add key backup automation
20. Consider HSM/KMS integration for production

---

## Conclusion

This is a **significant and well-architected security improvement** that modernizes SuperMQ's authentication to industry standards. The move from symmetric HMAC to asymmetric RSA with JWKS endpoint is the right direction.

### Key Strengths
- ✅ Proper secret isolation (private keys only in auth service)
- ✅ Automatic key rotation with graceful handling
- ✅ Standard-compliant JWKS endpoint
- ✅ Clean abstraction (KeyManager interface)
- ✅ Simplified service authentication (no gRPC client)

### Critical Issues to Address
- 🔴 JWKS cache race condition
- 🔴 Rotation handler crashes on error
- 🔴 Missing tests for KeyManager
- 🔴 Hardcoded JWKS URLs
- 🔴 Migration plan needed

### Overall Assessment
**Recommendation**: ✅ **Approve with required changes**

The architecture is sound, but several critical issues must be fixed:
1. Thread safety (JWKS cache)
2. Reliability (rotation error handling)
3. Testing (KeyManager tests)
4. Configuration (JWKS URL)
5. Documentation (migration guide)

After addressing these issues, this will be a major security and architecture improvement.

### Estimated Effort
- **Critical fixes**: 8-12 hours
- **High priority**: 16-20 hours
- **Medium priority**: 12-16 hours
- **Low priority**: 8-12 hours

### Risk Assessment
**Without fixes**: HIGH - Race conditions, rotation failures, migration issues

**With fixes**: LOW - Standard, well-understood authentication pattern

---

## Review Checklist

- [x] Code follows project conventions
- [x] Architecture is sound and well-designed
- [⚠️] Error handling needs improvement (rotation handler)
- [⚠️] Thread safety issues (JWKS cache)
- [ ] Tests are comprehensive (KeyManager untested)
- [ ] Integration tests needed
- [ ] Documentation needs to be added
- [x] Security improvement over previous approach
- [⚠️] Performance acceptable (RSA slower but cached)
- [x] Key rotation design is sound
- [⚠️] Migration plan needed (breaking change)
- [⚠️] Operational procedures needed (monitoring, recovery)

**Reviewed by**: Claude Code
**Date**: 2025-11-27
**Commit Range**: main...smq1672-token (5 commits)
