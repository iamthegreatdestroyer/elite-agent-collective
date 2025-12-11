# Elite Agent Collective - Testing Framework

Comprehensive testing strategies for the Elite Agent Collective system across all platforms and deployment scenarios.

## 🏗️ Testing Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Test Pyramid                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│                   E2E Tests (5%)                             │
│                  /            \                              │
│              /                    \                          │
│          Integration Tests (15%)                             │
│         /                      \                             │
│      /                            \                          │
│   Unit Tests (80%)                                           │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## 🧪 Unit Tests

Test individual components in isolation.

### Backend Unit Tests (Go)

```bash
# Run all unit tests
make test

# Run specific test
go test -v -run TestParseRequest ./internal/copilot

# Run with coverage
make test-coverage

# View coverage report
go tool cover -html=coverage.out
```

**Test File Structure** (`internal/agents/registry_test.go`):

```go
package agents

import (
    "testing"
)

func TestAgentRegistry(t *testing.T) {
    tests := []struct {
        name        string
        codename    string
        expectedTier int
        shouldExist bool
    }{
        {
            name:        "APEX exists in Tier 1",
            codename:    "APEX",
            expectedTier: 1,
            shouldExist: true,
        },
        {
            name:        "Invalid codename",
            codename:    "INVALID",
            expectedTier: 0,
            shouldExist: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agent := GetAgent(tt.codename)
            
            if tt.shouldExist && agent == nil {
                t.Errorf("Expected agent %s to exist", tt.codename)
            }
            
            if !tt.shouldExist && agent != nil {
                t.Errorf("Expected agent %s to not exist", tt.codename)
            }
            
            if agent != nil && agent.Tier != tt.expectedTier {
                t.Errorf("Expected tier %d, got %d", tt.expectedTier, agent.Tier)
            }
        })
    }
}

func TestGetAllAgents(t *testing.T) {
    agents := GetAllAgents()
    
    if len(agents) != 40 {
        t.Errorf("Expected 40 agents, got %d", len(agents))
    }
    
    // Verify each tier has correct number of agents
    tierCounts := make(map[int]int)
    for _, agent := range agents {
        tierCounts[agent.Tier]++
    }
    
    expectedCounts := map[int]int{
        1: 5,  // APEX, CIPHER, ARCHITECT, AXIOM, VELOCITY
        2: 12, // Specialists
        3: 2,  // NEXUS, GENESIS
        4: 1,  // OMNISCIENT
        5: 5,  // Domain specialists
        6: 5,  // Emerging tech
        7: 5,  // Human-centric
        8: 5,  // Enterprise
    }
    
    for tier, expected := range expectedCounts {
        if tierCounts[tier] != expected {
            t.Errorf("Tier %d: expected %d agents, got %d",
                tier, expected, tierCounts[tier])
        }
    }
}

func BenchmarkGetAgent(b *testing.B) {
    for i := 0; i < b.N; i++ {
        GetAgent("APEX")
    }
}
```

## 🔗 Integration Tests

Test interactions between components.

### Backend Integration Tests

```bash
# Run integration tests
make test-integration

# Run specific integration test
go test -v -run TestIntegrationCopilotRequest -tags integration ./tests/integration
```

**Example Integration Test** (`tests/integration/copilot_request_test.go`):

```go
// +build integration

package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestIntegrationCopilotAgentRequest(t *testing.T) {
    // Setup server
    router := setupTestRouter()
    server := httptest.NewServer(router)
    defer server.Close()

    // Create request
    requestBody := map[string]interface{}{
        "agent":     "APEX",
        "task":      "implement a rate limiter",
        "context":   "distributed system with 1000 rps",
        "language":  "go",
    }

    payload, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", server.URL+"/agent", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")

    // Execute
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    // Assert
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    if result["response"] == nil {
        t.Error("Response should not be empty")
    }

    // Verify response contains code examples
    response := result["response"].(string)
    if len(response) < 100 {
        t.Error("Response too short, expected substantial code example")
    }
}

func TestIntegrationMultiAgentCollaboration(t *testing.T) {
    router := setupTestRouter()
    server := httptest.NewServer(router)
    defer server.Close()

    agents := []string{"APEX", "ARCHITECT", "ECLIPSE"}
    
    for _, agent := range agents {
        requestBody := map[string]interface{}{
            "agent": agent,
            "task":  "design a distributed cache system",
        }

        payload, _ := json.Marshal(requestBody)
        req, _ := http.NewRequest("POST", server.URL+"/agent", bytes.NewBuffer(payload))
        req.Header.Set("Content-Type", "application/json")

        client := &http.Client{}
        resp, _ := client.Do(req)

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Agent %s failed: status %d", agent, resp.StatusCode)
        }
        resp.Body.Close()
    }
}

func TestIntegrationAuthenticationFlow(t *testing.T) {
    router := setupTestRouter()
    server := httptest.NewServer(router)
    defer server.Close()

    // Test without authentication
    req, _ := http.NewRequest("POST", server.URL+"/agent", bytes.NewBuffer([]byte("{}")))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, _ := client.Do(req)

    if resp.StatusCode != http.StatusUnauthorized {
        t.Errorf("Expected 401 without auth, got %d", resp.StatusCode)
    }

    // Test with valid token
    req.Header.Set("Authorization", "Bearer valid-token-xxx")
    resp, _ = client.Do(req)

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200 with auth, got %d", resp.StatusCode)
    }
}
```

## 🚀 End-to-End Tests

Test complete workflows across platforms.

### Platform Testing

#### VS Code Extension

```typescript
// tests/e2e/vs-code-extension.test.ts
import * as vscode from 'vscode';
import * as assert from 'assert';

suite('Elite Agent Collective Extension', () => {
    suite('Agent Invocation', () => {
        test('should invoke APEX agent', async () => {
            const extension = vscode.extensions.getExtension(
                'elite-agent-collective.elite-agents'
            );
            
            assert.ok(extension, 'Extension should be installed');
            
            await extension?.activate();
            assert.ok(extension?.isActive, 'Extension should be active');
        });

        test('should list all 40 agents', async () => {
            const agents = await vscode.commands.executeCommand(
                'elite-agents.listAgents'
            );
            
            assert.strictEqual((agents as any[]).length, 40);
        });

        test('should respond to agent query', async () => {
            const result = await vscode.commands.executeCommand(
                'elite-agents.invoke',
                {
                    agent: 'APEX',
                    task: 'implement binary search'
                }
            );
            
            assert.ok(result, 'Should return response');
            assert.ok((result as string).length > 0);
        });
    });

    suite('Multi-Agent Collaboration', () => {
        test('should coordinate APEX + ARCHITECT', async () => {
            const apexResponse = await vscode.commands.executeCommand(
                'elite-agents.invoke',
                { agent: 'APEX', task: 'design cache' }
            );
            
            const architectResponse = await vscode.commands.executeCommand(
                'elite-agents.invoke',
                { agent: 'ARCHITECT', task: 'review design' }
            );
            
            assert.ok(apexResponse);
            assert.ok(architectResponse);
        });
    });
});
```

#### GitHub.com Web

```javascript
// tests/e2e/github-web.test.js
describe('Elite Agents on GitHub.com', () => {
    beforeEach(async () => {
        // Navigate to GitHub.com
        await page.goto('https://github.com');
        
        // Open Copilot Chat
        await page.click('[aria-label="GitHub Copilot"]');
    });

    it('should show agent list in Copilot', async () => {
        // Type @ to trigger agent menu
        await page.type('[data-testid="chat-input"]', '@');
        
        // Wait for agents dropdown
        const agents = await page.$$('[data-agent-codename]');
        
        expect(agents.length).toBe(40);
    });

    it('should invoke agent and show response', async () => {
        // Type agent invocation
        await page.type(
            '[data-testid="chat-input"]',
            '@APEX implement a rate limiter'
        );
        
        // Send message
        await page.click('[aria-label="Send"]');
        
        // Wait for response
        await page.waitForSelector('[data-message-role="assistant"]');
        
        const response = await page.$eval(
            '[data-message-role="assistant"]',
            el => el.textContent
        );
        
        expect(response.length).toBeGreaterThan(100);
    });

    it('should handle multi-agent requests', async () => {
        const task = '@APEX design cache, @ARCHITECT review, @ECLIPSE test';
        
        await page.type('[data-testid="chat-input"]', task);
        await page.click('[aria-label="Send"]');
        
        // Verify all three agents responded
        const messages = await page.$$('[data-message-role="assistant"]');
        expect(messages.length).toBeGreaterThanOrEqual(3);
    });
});
```

#### JetBrains IDE

```java
// tests/e2e/JetbrainsTest.java
public class EliteAgentJetbrainsTest {
    private static EditorFixture editor;

    @BeforeAll
    public static void setUp() {
        editor = openIntelliJProject();
    }

    @Test
    public void testAgentMenuAppears() {
        // Open intention menu
        editor.invokeAction("ShowIntentionActions");
        
        // Agents should appear in menu
        MenuItemFixture agentMenu = editor.findMenuItem("Elite Agents");
        assertThat(agentMenu.isVisible()).isTrue();
    }

    @Test
    public void testInvokeAgent() {
        String selectedCode = "public int[] merge(int[] nums) { ... }";
        editor.select(selectedCode);
        
        // Invoke APEX agent
        editor.runAgent("APEX", "Optimize this merge algorithm");
        
        String suggestion = editor.getLastContextMenuAction();
        assertThat(suggestion).contains("time complexity");
    }

    @Test
    public void testAllAgentsAvailable() {
        List<String> agents = editor.listAvailableAgents();
        assertThat(agents).hasSize(40);
        assertThat(agents).contains("APEX", "CIPHER", "ARCHITECT");
    }
}
```

## 📊 Performance Testing

Test system performance under load.

```bash
# Load test with k6
k6 run tests/load/agent-load-test.js

# Results: RPS, latency p50/p95/p99, error rate
```

**k6 Load Test** (`tests/load/agent-load-test.js`):

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  vus: 100,  // Virtual users
  duration: '5m',
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95th percentile < 500ms
    http_req_failed: ['rate<0.01'],     // Error rate < 1%
  },
};

export default function () {
  const agents = ['APEX', 'CIPHER', 'ARCHITECT', 'TENSOR', 'FORTRESS'];
  const agent = agents[Math.floor(Math.random() * agents.length)];

  let response = http.post(
    `${__ENV.BASE_URL}/agent`,
    JSON.stringify({
      agent: agent,
      task: 'analyze this code: function sort(arr) { return arr.sort(); }',
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${__ENV.API_TOKEN}`,
      },
    }
  );

  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 1s': (r) => r.timings.duration < 1000,
    'has response text': (r) => r.body.length > 0,
  });
}
```

## 🔐 Security Testing

Test security properties of the system.

```bash
# Run security tests
make test-security

# OWASP ZAP scan
./run-zap-scan.sh http://localhost:8080

# Dependency scanning
go mod audit

# Code scanning
gosec ./...
```

## ✅ Test Coverage Requirements

| Component | Target Coverage |
|-----------|-----------------|
| agents/   | 90%+            |
| auth/     | 95%+            |
| copilot/  | 85%+            |
| memory/   | 80%+            |
| utils/    | 75%+            |
| **Total** | **85%+**        |

## 🚦 CI/CD Testing Pipeline

GitHub Actions runs tests on every commit:

```yaml
# .github/workflows/tests.yml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: make test-coverage
      - uses: codecov/codecov-action@v2

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
      redis:
        image: redis:7
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: npm install
      - run: npm run test:e2e
```

## 📈 Test Metrics Dashboard

Track testing metrics:

- **Coverage**: 85%+ overall
- **Pass Rate**: 99%+
- **Test Duration**: < 5 minutes total
- **Flaky Tests**: < 0.1%
- **Performance**: p95 latency < 500ms

---

**Last Updated**: December 11, 2025  
**Version**: 2.0.0
