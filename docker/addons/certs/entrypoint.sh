#!/bin/sh

# Simple OpenBao entrypoint for SuperMQ
set -e

echo "=== Starting OpenBao for SuperMQ ==="

# Start OpenBao in development mode in the background
bao server -dev \
  -dev-root-token-id="${BAO_DEV_ROOT_TOKEN_ID:-openbao-root-token}" \
  -dev-listen-address="0.0.0.0:8200" \
  -log-level=info &

BAO_PID=$!
echo "OpenBao started with PID: $BAO_PID"

# Simple wait for OpenBao to be ready
echo "Waiting for OpenBao to be ready..."
sleep 10

# Check if OpenBao is responding
for i in $(seq 1 10); do
  if curl -s http://localhost:8200/v1/sys/health >/dev/null 2>&1; then
    echo "OpenBao is ready!"
    break
  fi
  echo "Waiting... ($i/10)"
  sleep 2
done

# Setup OpenBao configuration
export BAO_ADDR=http://localhost:8200
export BAO_TOKEN="${BAO_DEV_ROOT_TOKEN_ID:-openbao-root-token}"

echo "Configuring OpenBao..."

# Enable required engines (ignore errors if already enabled)
bao auth enable approle 2>/dev/null || echo "AppRole already enabled"
bao secrets enable -path=pki pki 2>/dev/null || echo "PKI already enabled"

# Basic PKI setup
bao secrets tune -max-lease-ttl=87600h pki >/dev/null 2>&1 || true
bao write -field=certificate pki/root/generate/internal \
  common_name='SuperMQ Root CA' ttl=87600h >/dev/null 2>&1 || true

bao write pki/config/urls \
  issuing_certificates='http://localhost:8200/v1/pki/ca' \
  crl_distribution_points='http://localhost:8200/v1/pki/crl' >/dev/null 2>&1 || true

bao write pki/roles/supermq \
  allow_any_name=true enforce_hostnames=false allow_ip_sans=true \
  allow_localhost=true max_ttl=720h ttl=720h >/dev/null 2>&1 || true

# Create simple policy
cat > /tmp/policy.hcl << 'EOF'
path "pki/issue/supermq" { capabilities = ["create", "update"] }
path "pki/certs" { capabilities = ["list"] }
path "pki/cert/*" { capabilities = ["read"] }
path "pki/revoke" { capabilities = ["create", "update"] }
path "auth/token/renew-self" { capabilities = ["update"] }
path "auth/token/lookup-self" { capabilities = ["read"] }
EOF

bao policy write pki-policy /tmp/policy.hcl >/dev/null 2>&1 || true

# Create AppRole
bao write auth/approle/role/supermq \
  token_policies=pki-policy token_ttl=1h token_max_ttl=4h \
  bind_secret_id=true >/dev/null 2>&1 || true

# Set custom credentials if provided
if [ -n "$SMQ_CERTS_OPENBAO_APP_ROLE" ]; then
  bao write auth/approle/role/supermq/role-id role_id="$SMQ_CERTS_OPENBAO_APP_ROLE" >/dev/null 2>&1 || true
fi

if [ -n "$SMQ_CERTS_OPENBAO_APP_SECRET" ]; then
  bao write auth/approle/role/supermq/custom-secret-id secret_id="$SMQ_CERTS_OPENBAO_APP_SECRET" >/dev/null 2>&1 || true
fi

echo "OpenBao configuration completed successfully!"
echo "OpenBao is ready for SuperMQ on port 8200"

# Keep the process running
wait $BAO_PID
