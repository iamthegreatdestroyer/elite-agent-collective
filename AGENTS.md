# Elite Agent Collective - Developer Guide for GitHub Copilot

This file provides guidance for GitHub Copilot coding agent when working with the Elite Agent Collective repository.

## Repository Overview

The Elite Agent Collective is a comprehensive system of 40 specialized AI agents designed for GitHub Copilot. The repository consists of:

- **Agent Instructions**: Copilot instructions defining all 40 specialized agents in `.github/copilot-instructions.md`
- **Backend Service**: Go HTTP server that powers the GitHub Copilot Extension API
- **VS Code Integration**: Prompt files for VS Code user prompts
- **Testing Framework**: Python-based test suite for agent capabilities
- **Documentation**: Comprehensive docs for users and developers

## Technology Stack

### Backend (Go)
- **Language**: Go 1.21 or later (CI uses 1.21, go.mod specifies 1.24.10)
- **Framework**: Standard library HTTP server with custom middleware
- **Architecture**: Clean architecture with separation of concerns
- **Key Packages**:
  - `cmd/server`: Main entry point
  - `internal/agents`: Agent registry and handlers
  - `internal/copilot`: Copilot request/response handling
  - `internal/auth`: OIDC authentication and signature verification
  - `internal/memory`: Memory and retrieval systems
  - `internal/config`: Configuration management
  - `pkg/`: Public APIs and utilities

### Testing (Python)
- **Language**: Python 3.8+
- **Framework**: Custom test framework with pytest-style structure
- **Key Components**:
  - `tests/framework/`: Base test infrastructure
  - `tests/integration/`: Integration tests for agent collaboration
  - `tests/supreme_master_suite/`: Advanced testing scenarios

## Project Structure

```
elite-agent-collective/
├── .github/
│   ├── copilot-instructions.md    # Main Copilot instructions (40 agents)
│   └── workflows/                  # CI/CD pipelines
│       └── integration-tests.yml   # Go backend testing
├── backend/                        # Go HTTP server
│   ├── cmd/server/                 # Main entry point
│   ├── internal/                   # Private packages
│   │   ├── agents/                 # Agent registry and handlers
│   │   ├── auth/                   # Authentication (OIDC, signatures)
│   │   ├── config/                 # Configuration
│   │   ├── copilot/               # Copilot protocol types
│   │   └── memory/                # Memory and retrieval systems
│   ├── pkg/                        # Public packages
│   ├── tests/integration/          # Go integration tests
│   ├── Makefile                    # Build and test commands
│   ├── go.mod                      # Go module definition
│   └── Dockerfile                  # Container image
├── vscode-prompts/                 # VS Code integration
│   ├── ELITE_AGENT_COLLECTIVE.instructions.md
│   └── agents/                     # Individual agent files
├── tests/                          # Python test suite
│   ├── framework/                  # Test infrastructure
│   ├── integration/                # Integration tests
│   └── supreme_master_suite/       # Advanced test scenarios
├── docs/                           # Documentation
├── CONTRIBUTING.md                 # Contribution guidelines
└── README.md                       # User-facing README
```

## Development Workflow

### Backend Development

#### Prerequisites
- Go 1.21 or later (CI tests with 1.21, go.mod has 1.24.10)
- Docker (optional)

#### Common Commands
All commands should be run from the `backend/` directory:

```bash
# Install dependencies
make deps

# Run the server locally (default port 8080)
make run

# Build the binary
make build

# Format code
make fmt

# Lint code (uses golangci-lint if available, otherwise go vet)
make lint

# Run all unit tests
make test

# Run tests with coverage report
make test-coverage

# Run integration tests
make test-integration

# Run Copilot-specific tests
make test-copilot

# Run signature verification tests
make test-signature

# Run streaming response tests
make test-streaming

# Run all tests (unit + integration)
make test-all

# Run integration benchmarks
make test-bench

# Clean build artifacts
make clean

# Docker commands
make docker           # Build Docker image
make docker-run       # Run container
make compose-up       # Start with docker-compose
make compose-down     # Stop docker-compose
```

#### Code Style and Conventions

1. **Go Formatting**: Always run `make fmt` before committing
2. **Go Vet**: Code must pass `go vet` checks
3. **Package Organization**: Follow the clean architecture pattern
   - `cmd/`: Main applications
   - `internal/`: Private application code (not importable by other projects)
   - `pkg/`: Public libraries (importable by other projects)
4. **Error Handling**: Always check and handle errors explicitly
5. **Testing**: Write table-driven tests, use subtests with `t.Run()`
6. **Naming**: Use idiomatic Go naming conventions
   - Packages: lowercase, single word
   - Interfaces: `-er` suffix (e.g., `Handler`, `Authenticator`)
   - Constants: MixedCaps or ALL_CAPS for public constants

#### API Design

The backend follows the GitHub Copilot Extension API specification:

- **Health Check**: `GET /health` - Returns service health status
- **List Agents**: `GET /agents` - Returns all registered agents
- **Get Agent**: `GET /agents/{codename}` - Returns specific agent info
- **Agent Query**: `POST /agent` - Processes agent queries with Copilot protocol

Request/Response format follows the Copilot protocol defined in `internal/copilot/`.

### Python Testing

The Python test suite validates agent capabilities and interactions:

```bash
# Run from repository root or tests/ directory
python tests/supreme_master_suite/run_supreme_suite.py
```

Test files follow pytest-style structure with descriptive test names.

## Testing Guidelines

### Unit Tests (Go)
- Place tests in the same package as the code (`_test.go` suffix)
- Use table-driven tests for multiple scenarios
- Mock external dependencies using interfaces
- Aim for 80%+ coverage on critical paths

### Integration Tests (Go)
- Located in `backend/tests/integration/`
- Use build tag: `// +build integration`
- Test full request/response cycles
- Verify signature verification and authentication flows

### Test Naming
- Unit tests: `Test<FunctionName>` (e.g., `TestParseRequest`)
- Integration tests: `TestIntegration<Feature>` (e.g., `TestIntegrationCopilotRequest`)
- Use subtests for variants: `t.Run("description", func(t *testing.T) { ... })`

## Key Concepts

### Agent Registry
All 40 agents are registered in `internal/agents/registry.go`. Each agent has:
- ID (01-40)
- Codename (e.g., APEX, CIPHER)
- Tier (1-8)
- Specialty description
- Philosophy statement
- Capabilities list
- Directives for Copilot

### Authentication
- **OIDC**: OpenID Connect token validation (stub implementation in `internal/auth/oidc.go`)
- **Signature Verification**: GitHub request signature validation in `internal/auth/signature.go`
- **Middleware**: Authentication middleware in `internal/auth/middleware.go`

### Memory System
The memory system (`internal/memory/`) implements:
- Sub-linear retrieval for efficient memory access
- ReMem control loop for memory management
- Experience tracking across agent interactions

## CI/CD Pipeline

### GitHub Actions
Workflow: `.github/workflows/integration-tests.yml`

Runs on:
- Push to `main` branch (backend changes)
- Pull requests to `main` (backend changes)
- Manual trigger via `workflow_dispatch`

Jobs:
1. **integration**: Runs unit tests, integration tests, and benchmarks
2. **lint**: Runs `go vet` and `go fmt` checks
3. **build**: Builds the binary and uploads as artifact

## Common Tasks

### Adding a New Agent
1. Add agent definition to `internal/agents/registry.go`
2. Create handler in `internal/agents/handlers/` if needed
3. Update `.github/copilot-instructions.md` with agent details
4. Create VS Code prompt file in `vscode-prompts/agents/`
5. Update README.md agent registry table
6. Add tests for agent behavior
7. Update documentation

### Modifying API Endpoints
1. Update handler in `internal/agents/handler.go` or specific handler
2. Ensure proper error handling and validation
3. Update tests in `*_test.go` files
4. Run integration tests: `make test-integration`
5. Update API documentation if needed

### Fixing a Bug
1. Write a failing test that reproduces the bug
2. Fix the bug with minimal changes
3. Ensure all tests pass: `make test-all`
4. Run linter: `make lint`
5. Format code: `make fmt`

### Performance Optimization
1. Identify bottleneck with profiling
2. Write benchmark test in `*_test.go`
3. Run benchmarks: `make test-bench`
4. Implement optimization
5. Re-run benchmarks to verify improvement
6. Ensure no regressions in functionality tests

## Security Considerations

- All authentication code is in `internal/auth/`
- OIDC token validation is required for production
- Request signatures must be verified for GitHub Copilot requests
- Never log sensitive data (tokens, credentials)
- Follow principle of least privilege for permissions
- Validate all input data before processing

## Documentation

- **User Docs**: `docs/` directory with user-facing guides
- **API Reference**: `docs/api-reference/` for backend API
- **Developer Guide**: `docs/developer-guide/` for contributor information
- **Inline Comments**: Document complex logic, not obvious code

## Deployment

### Docker Deployment
```bash
# Build image
docker build -t elite-agent-collective backend/

# Run container
docker run -p 8080:8080 elite-agent-collective
```

### Docker Compose
```bash
cd backend/
docker-compose up -d
```

Environment variables:
- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Troubleshooting

### Build Issues
- Ensure Go 1.21 or later is installed: `go version`
- Clean and rebuild: `make clean && make build`
- Check dependencies: `make deps`

### Test Failures
- Run specific test: `go test -v -run TestName ./path/to/package`
- Check test output for error details
- Verify test data and fixtures

### Server Issues
- Check logs for error messages
- Verify port 8080 is not in use
- Ensure environment variables are set correctly
- Test health endpoint: `curl http://localhost:8080/health`

## Resources

- [Contributing Guide](CONTRIBUTING.md)
- [Full Documentation](docs/README.md)
- [GitHub Copilot Extension Docs](https://docs.github.com/en/copilot/building-copilot-extensions)
- [Go Best Practices](https://go.dev/doc/effective_go)

## Support

- GitHub Issues: Report bugs and feature requests
- GitHub Discussions: Ask questions and share ideas
- Documentation: Check docs/ directory first

---

**Note for Copilot**: When working on this repository:
1. Always run tests after making changes (`make test-all` in backend/)
2. Format code before committing (`make fmt`)
3. Follow Go idioms and the existing code style
4. Update tests when modifying behavior
5. Check CI status on pull requests
6. Update documentation for user-facing changes
7. Use appropriate agents from the Elite Agent Collective when asking for help (see `.github/copilot-instructions.md`)

*"The collective intelligence of specialized minds exceeds the sum of their parts."*
