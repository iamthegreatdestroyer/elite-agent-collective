# OIDC Authentication Setup

This guide explains how to configure OpenID Connect (OIDC) authentication for the Elite Agent Collective backend.

## Overview

The backend supports OIDC authentication for protected endpoints. When enabled, it validates JWT tokens using the OIDC provider's public keys (JWKS). The implementation includes:

- JWT token validation with RS256 signature verification
- JWKS endpoint fetching with 1-hour caching
- Issuer, audience, and expiration validation
- Support for GitHub Actions OIDC tokens

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OIDC_ISSUER` | No | `https://token.actions.githubusercontent.com` | OIDC issuer URL |
| `OIDC_CLIENT_ID` | Yes* | - | Client ID for token audience validation |
| `OIDC_CLIENT_SECRET` | No | - | Client secret (if required by provider) |

**\* Authentication is only enabled when `OIDC_CLIENT_ID` is set. If not set, all endpoints are publicly accessible.**

## Protected vs Public Endpoints

When OIDC is enabled:

| Endpoint | Authentication |
|----------|---------------|
| `GET /health` | Public |
| `GET /agents` | Public |
| `GET /agents/{codename}` | Public |
| `POST /agents/{codename}/invoke` | **Required** |
| `POST /copilot` | **Required** |

## Local Development

For local development without authentication, simply don't set the `OIDC_CLIENT_ID` environment variable:

```bash
# Start server without authentication
make run

# Or with docker-compose
docker-compose up
```

All endpoints will be accessible without tokens.

## GitHub Actions OIDC Setup

To use GitHub Actions OIDC tokens:

### 1. Configure Environment Variables

Set these environment variables in your deployment:

```bash
export OIDC_ISSUER="https://token.actions.githubusercontent.com"
export OIDC_CLIENT_ID="your-repository-owner/your-repository-name"
```

### 2. GitHub Actions Workflow

In your GitHub Actions workflow, request an OIDC token:

```yaml
name: Deploy and Test
on: [push]

permissions:
  id-token: write  # Required for OIDC token
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Get OIDC Token
        id: get-token
        run: |
          TOKEN=$(curl -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
            "$ACTIONS_ID_TOKEN_REQUEST_URL&audience=${{ github.repository_owner }}/${{ github.repository }}" \
            | jq -r '.value')
          echo "token=$TOKEN" >> $GITHUB_OUTPUT
      
      - name: Call Protected Endpoint
        run: |
          curl -X POST \
            -H "Authorization: Bearer ${{ steps.get-token.outputs.token }}" \
            -H "Content-Type: application/json" \
            -d '{"messages": [{"role": "user", "content": "Hello"}]}' \
            https://your-api.example.com/agents/APEX/invoke
```

## Testing Authentication

### Test with curl (no auth)

When OIDC is not configured:

```bash
# All endpoints are accessible
curl http://localhost:8080/agents
curl http://localhost:8080/agents/APEX
curl -X POST http://localhost:8080/agents/APEX/invoke \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello"}]}'
```

### Test with curl (auth enabled)

When OIDC is configured:

```bash
# Public endpoints work without token
curl http://localhost:8080/health
curl http://localhost:8080/agents

# Protected endpoints require Bearer token
curl -X POST http://localhost:8080/agents/APEX/invoke \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello"}]}'

# Without token, returns 401
curl -X POST http://localhost:8080/agents/APEX/invoke \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello"}]}'
# Response: Authorization header required
```

### Generate Test Token

For testing purposes, you can generate a test JWT using OpenSSL and a scripting language. Here's an example using Python:

```bash
# Install PyJWT if not already installed
pip install PyJWT cryptography

# Generate a test token with a test RSA key pair
python3 -c "
import jwt
import datetime
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.backends import default_backend

# Generate a test RSA key pair
private_key = rsa.generate_private_key(
    public_exponent=65537,
    key_size=2048,
    backend=default_backend()
)

# Create the token
payload = {
    'sub': 'test-user',
    'iss': 'https://your-issuer.example.com',
    'aud': 'your-client-id',
    'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=1)
}

token = jwt.encode(
    payload,
    private_key,
    algorithm='RS256',
    headers={'kid': 'test-key-id'}
)

print(token)
"
```

Note: For production testing, use tokens from your actual OIDC provider.

## JWKS Caching

The backend automatically caches public keys from the JWKS endpoint:

- **Cache TTL**: 1 hour
- **Automatic refresh**: Keys are refreshed when cache expires
- **Thread-safe**: Multiple concurrent requests share the cache

This reduces load on the OIDC provider and improves token validation performance.

## Troubleshooting

### "Authorization header required"

The request is missing the Authorization header. Add:
```
Authorization: Bearer YOUR_TOKEN
```

### "Invalid authorization header format"

The Authorization header must use the Bearer scheme:
```
Authorization: Bearer YOUR_TOKEN
```

### "Invalid token"

Common causes:
- Token has expired (`exp` claim in the past)
- Token issuer doesn't match `OIDC_ISSUER`
- Token audience doesn't match `OIDC_CLIENT_ID`
- Token signature is invalid
- JWKS endpoint is unreachable

Check server logs for detailed error messages.

### "token missing key ID (kid)"

The JWT must include a `kid` (key ID) header that matches a key in the JWKS.

### JWKS endpoint errors

Ensure the OIDC issuer's discovery endpoint is accessible:
```bash
curl https://your-issuer.example.com/.well-known/openid-configuration
```

The response should include a `jwks_uri` field pointing to the JWKS endpoint.

## Security Considerations

1. **Always use HTTPS** in production
2. **Set appropriate token expiration** (shorter is better)
3. **Validate audience** to prevent token confusion attacks
4. **Monitor for failed authentication attempts**
5. **Rotate keys regularly** on the OIDC provider

## Example Deployment

### Docker

```bash
docker run -p 8080:8080 \
  -e OIDC_ISSUER="https://token.actions.githubusercontent.com" \
  -e OIDC_CLIENT_ID="your-org/your-repo" \
  elite-agent-collective
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elite-agent-collective
spec:
  template:
    spec:
      containers:
        - name: server
          image: elite-agent-collective:latest
          env:
            - name: OIDC_ISSUER
              value: "https://token.actions.githubusercontent.com"
            - name: OIDC_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: oidc-config
                  key: client-id
```
