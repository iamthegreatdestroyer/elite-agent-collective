# Developer Guide

Complete guide for developers who want to understand, extend, or contribute to the Elite Agent Collective.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP Request (POST /agent)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
├→ Authentication (OIDC / Signature Verification)                │
│  └─ Valid? ✓ Continue : ✗ Return 401                          │
│                                                                 │
├→ Route Handler (handler.go)                                    │
│  └─ Extract agent codename from request body                   │
│                                                                 │
├→ Agent Registry (registry.go)                                  │
│  └─ Look up handler for agent codename                         │
│                                                                 │
├→ Agent Handler (handlers/*)                                    │
│  ├─ GetInfo() → Return agent metadata                          │
│  └─ Handle() → Process request via LLM/logic                   │
│                                                                 │
├→ MNEMONIC Memory System (memory/)                              │
│  ├─ Retrieve past experiences (ReMem RETRIEVE phase)           │
│  ├─ Augment context with learned strategies                    │
│  └─ Store new experience with fitness score (ReMem EVOLVE)     │
│                                                                 │
├→ Response Construction                                         │
│  └─ Format response JSON with agent response + metadata        │
│                                                                 │
└─→ HTTP Response (200 OK with JSON body)                        │
    └─ Return to client/Copilot                                  │
```

## Directory Structure

```
elite-agent-collective/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point, routes initialization
│   │
│   ├── internal/
│   │   ├── agents/
│   │   │   ├── agents.go            # Agent registration & loading (REFACTORED)
│   │   │   ├── agent_loader.go      # YAML/Markdown parser (NEW)
│   │   │   ├── registry.go          # Agent registry & lookup
│   │   │   ├── handler.go           # HTTP request handler
│   │   │   ├── handlers/
│   │   │   │   ├── apex.go          # APEX agent implementation
│   │   │   │   └── base.go          # Base handler for other agents
│   │   │   ├── *_test.go            # Agent tests
│   │   │   └── handlers/*_test.go   # Handler tests
│   │   │
│   │   ├── auth/
│   │   │   ├── oidc.go              # OIDC token validation
│   │   │   ├── signature.go         # GitHub request signature verification
│   │   │   ├── middleware.go        # Authentication middleware
│   │   │   └── *_test.go            # Auth tests
│   │   │
│   │   ├── config/
│   │   │   └── config.go            # Configuration management
│   │   │
│   │   ├── copilot/
│   │   │   ├── request.go           # GitHub Copilot request types
│   │   │   ├── response.go          # GitHub Copilot response types
│   │   │   └── *_test.go            # Copilot protocol tests
│   │   │
│   │   └── memory/
│   │       ├── experience.go        # Experience storage types
│   │       ├── remem_loop.go        # ReMem control loop
│   │       ├── sublinear_retriever.go # Sub-linear retrieval (Bloom, LSH, HNSW)
│   │       ├── advanced_structures.go # Phase 1 structures
│   │       ├── agent_aware_structures.go # Phase 2 structures
│   │       └── *_test.go            # Memory tests
│   │
│   ├── pkg/
│   │   ├── copilot/
│   │   │   ├── types.go             # Public Copilot types
│   │   │   └── manifest.go          # Copilot manifest generation
│   │   │
│   │   └── models/
│   │       ├── agent.go             # Agent model (EXTENDED)
│   │       └── memory.go            # Memory types
│   │
│   ├── tests/
│   │   ├── integration/
│   │   │   ├── agent_test.go        # Full agent tests
│   │   │   ├── handler_test.go      # Handler tests
│   │   │   └── memory_test.go       # Memory system tests
│   │   │
│   │   └── fixtures/
│   │       ├── agents.json          # Sample agent data
│   │       └── requests.json        # Sample requests
│   │
│   ├── go.mod                       # Go dependencies
│   ├── Makefile                     # Build commands
│   ├── Dockerfile                   # Container image
│   └── docker-compose.yml           # Container orchestration
│
├── .github/
│   └── agents/                      # Agent definition files
│       ├── APEX.agent.md
│       ├── CIPHER.agent.md
│       └── ... (40 total)
│
├── docs/                            # Documentation
│   ├── QUICK_START.md
│   ├── INSTALLATION.md
│   ├── API_REFERENCE.md
│   └── DEVELOPER_GUIDE.md (this file)
│
└── tests/                           # Python test suite
    ├── framework/
    ├── integration/
    └── supreme_master_suite/
```

## Core Modules

### 1. Agent Loader (agent_loader.go)

**Purpose:** Parse `.agent.md` files and populate Agent model

**Key Functions:**

```go
// Load single agent from file
func LoadAgentFromFile(filePath string) (*models.Agent, error)

// Load all agents from directory
func LoadAllAgentsFromDirectory(agentsDir string) ([]models.Agent, error)

// Parse YAML frontmatter
func parseFrontmatter(content string) (*AgentFileMetadata, string, error)

// Extract sections from markdown
func extractPhilosophy(content string) string
func extractDirectives(content string) []string
func extractExamples(content string) []string
func extractCollaborators(content string) []string
func ValidateAgentID(idStr string) (int, error)
```

**Data Flow:**

```
.agent.md file
    ↓
LoadAgentFromFile(path)
    ↓
parseFrontmatter() → extract YAML + markdown
    ↓
Apply extractPhilosophy, extractDirectives, etc.
    ↓
Populate models.Agent struct
    ↓
Return fully populated Agent
```

### 2. Agent Registry (registry.go)

**Purpose:** Central lookup and registration of all agents

**Key Components:**

```go
// Registry holds all agents, thread-safe
type Registry struct {
    mu       sync.RWMutex
    handlers map[string]AgentHandler
}

// Register an agent
func (r *Registry) Register(handler AgentHandler) error

// Get agent info
func (r *Registry) GetInfo(codename string) (*models.Agent, error)

// Get all agents
func (r *Registry) ListAgents() ([]models.Agent, error)

// Look up handler
func (r *Registry) Get(codename string) (AgentHandler, error)
```

**Thread Safety:**

- Uses `sync.RWMutex` for concurrent access
- Multiple goroutines can read simultaneously
- Only one writer at a time

### 3. Agent Handler (handler.go)

**Purpose:** HTTP request routing to agents

**Key Functions:**

```go
// Handle HTTP requests
func (h *AgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)

// Extract agent codename from request
func extractAgentCodename(body []byte) (string, error)

// Route to specific agent handler
func (h *AgentHandler) routeToAgent(w http.ResponseWriter, r *http.Request, agent string)
```

**Request Flow:**

```
POST /agent (JSON body with agent + messages)
    ↓
extractAgentCodename() → Get agent codename
    ↓
registry.Get(codename) → Get handler
    ↓
handler.Handle(req) → Process via agent
    ↓
Format response as JSON
    ↓
Write HTTP 200 response
```

### 4. Base Agent Handler (handlers/base.go)

**Purpose:** Generic agent handler for most agents

```go
type BaseAgent struct {
    definition models.Agent
    // ... other fields
}

func NewBaseAgent(def models.Agent) AgentHandler {
    return &BaseAgent{definition: def}
}

func (a *BaseAgent) GetInfo() *models.Agent {
    return &a.definition
}

func (a *BaseAgent) Handle(req *copilot.Request) (*copilot.Response, error) {
    // Call AI/LLM with agent's philosophy and directives
    // Return formatted response
}
```

### 5. Agent Model (pkg/models/agent.go)

**Current Structure:**

```go
type Agent struct {
    // Original fields
    ID         string   // "01" - "40"
    Codename   string   // "APEX", "CIPHER", etc.
    Tier       int      // 1-8
    Specialty  string   // Primary expertise
    Philosophy string   // Core principle
    Directives []string // Core capabilities

    // New fields (from .agent.md)
    Name          string   // Full name
    Keywords      []string // Search keywords
    Examples      []string // Invocation examples
    Collaborators []string // Agent references
    Category      string   // Tier category
    MarkdownPath  string   // Source file path (json:"-")
}
```

### 6. MNEMONIC Memory System (memory/)

**Purpose:** Store and retrieve agent experiences for continuous learning

**Key Components:**

- **Bloom Filter** - O(1) exact matching
- **LSH Index** - O(1) approximate nearest neighbor search
- **HNSW Graph** - O(log n) semantic search
- **Count-Min Sketch** - Frequency estimation
- **Cuckoo Filter** - Set membership with deletion
- **Product Quantizer** - Embedding compression
- **Advanced Agent-Aware Structures** - Cross-agent collaboration tracking

**ReMem Control Loop:**

```
RETRIEVE
    ↓
Query Bloom Filter → LSH → HNSW for relevant past experiences
    ↓
THINK
    ↓
Augment current request with retrieved strategies
    ↓
ACT
    ↓
Execute agent with memory-enhanced context
    ↓
REFLECT
    ↓
Evaluate success and compute fitness score
    ↓
EVOLVE
    ↓
Store new experience with fitness metadata
Promote successful experiences as breakthroughs
```

## Development Workflow

### Adding a New Feature

**Step 1: Create Feature Branch**

```bash
git checkout -b feature/my-feature
cd backend
```

**Step 2: Write Tests First**

```go
// internal/agents/feature_test.go
func TestMyFeature(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

**Step 3: Implement Feature**

```go
// internal/agents/feature.go
func MyFeature() {
    // Implementation
}
```

**Step 4: Run Tests**

```bash
make test
# or:
go test -v ./...
```

**Step 5: Format Code**

```bash
make fmt
# or:
go fmt ./...
```

**Step 6: Build & Verify**

```bash
make build
make run  # In another terminal
curl http://localhost:8080/health
```

**Step 7: Commit**

```bash
git add .
git commit -m "feat(module): description of change"
git push origin feature/my-feature
```

**Step 8: Create Pull Request**

- Title: `feat(module): description`
- Description: What changed and why
- Link any related issues

### Adding a New Agent

**Step 1: Create Agent File**

```bash
cp .github/agents/APEX.agent.md .github/agents/NEWAGENT.agent.md
```

**Step 2: Update YAML Frontmatter**

```yaml
---
name: New Agent Name
description: Brief description
codename: NEWAGENT
tier: 1
id: "41" # Next available ID
category: "Category"
keywords:
  - keyword1
  - keyword2
---
```

**Step 3: Update Markdown Sections**

- Philosophy statement
- Primary Function
- Capabilities
- Methodology
- Examples
- Integration Notes

**Step 4: Test Loading**

```bash
cd backend
go test -v ./internal/agents/...
# Should show: Loaded 41 agents from ...
```

**Step 5: Verify via API**

```bash
go run ./cmd/server
curl http://localhost:8080/agents/NEWAGENT
```

**Step 6: Commit & Push**

```bash
git add .github/agents/NEWAGENT.agent.md
git commit -m "feat(agents): add NEWAGENT"
```

## Testing

### Unit Tests

```bash
# Run all unit tests
cd backend
go test -v ./...

# Run specific package
go test -v ./internal/agents/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests only
cd backend
go test -v -tags=integration ./tests/integration/...

# Run all tests including integration
make test-all
```

### Test Structure

```go
// Table-driven tests
func TestRegistry(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"test case 1", "input", "expected", false},
        {"test case 2 error", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Implement test
        })
    }
}

// Subtests
func TestComplex(t *testing.T) {
    t.Run("scenario 1", func(t *testing.T) {
        // Test scenario 1
    })

    t.Run("scenario 2", func(t *testing.T) {
        // Test scenario 2
    })
}
```

## Performance Considerations

### Agent Loading

- **Time:** ~287ms to load all 40 agents
- **File Size:** ~170-320 lines per agent
- **Complexity:** O(n) linear scan of directory

### Agent Lookup

- **Time:** ~0.02ms (O(1) map lookup)
- **Space:** ~1MB for all agent metadata

### Memory System

| Operation           | Complexity     | Time   |
| ------------------- | -------------- | ------ |
| Bloom Filter lookup | O(1)           | ~0.1μs |
| LSH search          | O(1) expected  | ~0.5μs |
| HNSW navigation     | O(log n)       | ~1-2μs |
| Store experience    | O(1) amortized | ~2-5μs |

### HTTP Request Handling

- **Routing:** <1ms
- **Authentication:** <5ms (if enabled)
- **Agent invocation:** 200-2000ms (depends on LLM)
- **Response formatting:** <5ms

## Security

### Code Review Checklist

- [ ] No hardcoded secrets (keys, tokens, passwords)
- [ ] All inputs validated and sanitized
- [ ] Errors don't leak sensitive information
- [ ] Proper authentication/authorization checks
- [ ] Follows principle of least privilege
- [ ] Uses secure defaults (e.g., https in production)
- [ ] No path traversal vulnerabilities
- [ ] SQL injection protection (if applicable)
- [ ] CSRF protection (if applicable)

### Secure Development

```go
// ✓ Good: Validate input
func ProcessAgent(codename string) error {
    if !isValidCodename(codename) {
        return fmt.Errorf("invalid codename")
    }
    // Process...
}

// ✗ Bad: No validation
func ProcessAgent(codename string) error {
    // Process directly without checks
}

// ✓ Good: Don't log sensitive data
logger.Info("Processing request", "agent", agent)

// ✗ Bad: Logs token
logger.Info("Request", "token", bearerToken)

// ✓ Good: Use environment variables for secrets
apiKey := os.Getenv("ANTHROPIC_API_KEY")

// ✗ Bad: Hardcoded secret
apiKey := "sk-123456..."
```

## Common Tasks

### Debug Agent Loading

```bash
# Run with logging
LOG_LEVEL=debug go run ./cmd/server

# Check agent directory
ls -la ../.github/agents/ | wc -l  # Should be 40

# Check YAML parsing
cat .github/agents/APEX.agent.md | head -20

# Test YAML manually (requires yq)
yq eval '.codename' .github/agents/APEX.agent.md
```

### Debug HTTP Requests

```bash
# Enable request logging
LOG_LEVEL=debug go run ./cmd/server

# Capture traffic with curl
curl -v -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{"agent": "APEX", "messages": [...]}'

# Use tcpdump to monitor network
sudo tcpdump -i lo port 8080
```

### Profile Performance

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Benchmark
go test -bench=. -benchmem ./...
```

## Deployment

### Local Development

```bash
cd backend
make run
```

### Docker

```bash
# Build
make docker

# Run
make docker-run

# Or with compose
make compose-up
```

### Kubernetes

Deploy using Helm charts or Kustomize (see deployment docs).

### CI/CD

GitHub Actions pipeline in `.github/workflows/` automatically:

- Runs tests on push/PR
- Builds Docker image
- Pushes to registry (if configured)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed contribution guidelines.

## Resources

- **Go Documentation** - https://golang.org/doc/
- **HTTP Package** - https://golang.org/pkg/net/http/
- **YAML Package** - https://github.com/go-yaml/yaml
- **Testing** - https://golang.org/pkg/testing/
- **Benchmarking** - https://golang.org/pkg/testing/#hdr-Benchmarks

## Support

- **Code Review** - GitHub Issues/PRs
- **Questions** - GitHub Discussions
- **Documentation** - See docs/ directory
