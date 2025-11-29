# Integration Tests

This directory contains end-to-end integration tests for the Elite Agent Collective backend.

## Test Categories

### 1. Copilot Webhook Tests (`copilot_webhook_test.go`)
Tests the `/copilot` endpoint that handles GitHub Copilot requests:
- Single agent invocation
- Multi-agent invocation
- Unknown agent handling
- Malformed request handling
- Empty message handling
- Response format validation
- Default agent fallback
- Conversation context handling

### 2. Agent Invocation Tests (`agent_invocation_test.go`)
Tests all 40 agents can be invoked correctly:
- All agents invocable via `/agents/{codename}/invoke`
- Agent response contains identity
- Agent collaboration (sequential invocations)
- Agent context awareness
- Agent invocation via `/copilot` endpoint
- Agent count verification (40 agents)

### 3. Authentication Tests (`auth_test.go`)
Tests OIDC/JWT authentication flow:
- Valid token authentication
- Missing token handling
- Invalid token format handling
- Webhook signature validation
- Public endpoint bypass
- Protected endpoint enforcement
- Case-insensitive bearer prefix

### 4. API Endpoint Tests (`api_test.go`)
Tests all REST endpoints:
- `GET /health` - Health check
- `GET /agents` - List all agents
- `GET /agents/{codename}` - Get specific agent
- `POST /agents/{codename}/invoke` - Invoke agent
- `POST /copilot` - Copilot webhook
- Content negotiation
- Validation errors

### 5. Request/Response Format Tests (`format_test.go`)
Verifies Copilot-compatible message formats:
- Request parsing variations
- Response format validation
- Streaming response handling
- Error response format
- Message role parsing
- Agent mention formats
- Large request handling

### 6. Performance Tests (`performance_test.go`)
Basic performance benchmarks:
- Agent invocation benchmark
- Concurrent requests benchmark
- List agents benchmark
- Get agent benchmark
- Health check benchmark
- Response time tests (< 500ms)
- Concurrent request handling
- Memory under load test

## Running Tests

### Prerequisites
- Go 1.21 or later
- All dependencies installed (`go mod download`)

### Run Integration Tests
```bash
# From the backend directory
make test-integration

# Or directly with go test
go test -v -tags=integration ./tests/integration/...
```

### Run All Tests (Unit + Integration)
```bash
make test-all
```

### Run Integration Benchmarks
```bash
make test-bench

# Or with custom benchmark options
go test -v -tags=integration -bench=. -benchmem ./tests/integration/...
```

### Run Specific Test Categories
```bash
# Run only copilot webhook tests
go test -v -tags=integration -run TestCopilotWebhook ./tests/integration/...

# Run only auth tests
go test -v -tags=integration -run TestOIDC ./tests/integration/...

# Run only agent tests
go test -v -tags=integration -run TestAllAgents ./tests/integration/...
```

## Test Fixtures

Test fixtures are located in `tests/fixtures/`:

```
tests/fixtures/
├── copilot_requests/       # Sample Copilot request payloads
│   ├── single_agent.json
│   ├── multi_agent.json
│   ├── with_context.json
│   ├── malformed.json
│   ├── empty_message.json
│   └── no_agent.json
├── expected_responses/     # Expected response structures
│   ├── apex_response.json
│   └── error_response.json
└── auth/                   # Authentication test data
    ├── valid_token.json
    └── expired_token.json
```

## CI/CD Integration

Integration tests run automatically on:
- Push to `main` branch
- Pull requests to `main` branch

See `.github/workflows/integration-tests.yml` for the CI configuration.

## Writing New Tests

1. Add test file with `//go:build integration` build tag
2. Use the `integration` package
3. Use `getTestServerURL()` to get the test server URL
4. Follow existing test patterns for consistency

Example:
```go
//go:build integration
// +build integration

package integration

import (
    "net/http"
    "testing"
)

func TestNewFeature(t *testing.T) {
    resp, err := http.Get(getTestServerURL() + "/new-endpoint")
    if err != nil {
        t.Fatalf("failed to make request: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected status 200, got %d", resp.StatusCode)
    }
}
```

## Test Coverage

Current test coverage includes:
- ✅ All 40 agents invocable
- ✅ Copilot webhook endpoint
- ✅ Authentication middleware
- ✅ All public API endpoints
- ✅ Request/response format validation
- ✅ Performance benchmarks
