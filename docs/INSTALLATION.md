# Installation & Setup Guide

## Prerequisites

Before installing the Elite Agent Collective, ensure you have:

- **Go 1.21+** - [Download from golang.org](https://golang.org/dl/)
- **Git** - For version control and cloning the repository
- **Docker** (optional) - For containerized deployment
- **Python 3.8+** (optional) - For running test suite

### Verify Prerequisites

```bash
# Check Go version
go version
# Should show: go version go1.21.x or higher

# Check Git is installed
git --version

# Check Docker (optional)
docker --version
```

## Getting the Code

### Option 1: Clone from GitHub

```bash
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git
cd elite-agent-collective
```

### Option 2: Download as ZIP

1. Visit [GitHub repository](https://github.com/iamthegreatdestroyer/elite-agent-collective)
2. Click "Code" → "Download ZIP"
3. Extract the ZIP file
4. Open terminal in the extracted directory

## Backend Setup

### Step 1: Navigate to Backend Directory

```bash
cd backend
```

### Step 2: Install Dependencies

```bash
go mod download
# or use the Makefile
make deps
```

### Step 3: Verify Installation

```bash
go mod tidy
go mod verify
```

### Step 4: Build the Server

```bash
make build
# or manually:
go build -v -o bin/server ./cmd/server
```

### Step 5: Run the Server

```bash
# Using Makefile
make run

# Or manually:
go run ./cmd/server

# Or run the compiled binary
./bin/server
# On Windows:
.\bin\server.exe
```

**Expected Output:**

```
2025/12/11 10:00:00 Starting Elite Agent Collective backend server
2025/12/11 10:00:00 Loaded 40 agents from /path/to/.github/agents
2025/12/11 10:00:00 Server listening on :8080
```

### Step 6: Verify Server is Running

In a new terminal:

```bash
# Health check endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","agents_loaded":40,"timestamp":"2025-12-11T10:00:00Z"}

# List all agents
curl http://localhost:8080/agents

# Get specific agent
curl http://localhost:8080/agents/APEX
```

## Docker Setup (Optional)

### Build Docker Image

```bash
cd backend
docker build -t elite-agent-collective:latest .
```

### Run Docker Container

```bash
docker run -p 8080:8080 elite-agent-collective:latest
```

### Using Docker Compose

```bash
cd backend
docker-compose up -d
```

**Verify container is running:**

```bash
docker logs $(docker ps -q --filter ancestor=elite-agent-collective:latest)
```

## VS Code Integration Setup

### Step 1: Install VS Code Extension

1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X / Cmd+Shift+X)
3. Search for "Elite Agent Collective"
4. Click Install

### Step 2: Configure Extension

1. Open VS Code Settings (Ctrl+, / Cmd+,)
2. Search for "elite.agent"
3. Configure:
   - Backend URL: `http://localhost:8080`
   - API Key (if authentication enabled)
   - Preferred tier for auto-complete

### Step 3: Use in VS Code

In any file, type `@` to see agent suggestions:

```
@APEX implement a new feature
@CIPHER review code for security issues
@ARCHITECT design a new system
```

## GitHub Copilot Integration

### Prerequisites

- GitHub Copilot subscription or free trial
- VS Code with GitHub Copilot extension

### Step 1: Enable Copilot Extension API

In VS Code settings, search for `github.copilot.enable` and ensure it's checked.

### Step 2: Configure Agent Backend

The GitHub Copilot extension will automatically discover agents from:

- Local VS Code agent definitions
- Remote agent server (if configured in settings)

### Step 3: Use with Copilot

```python
# In any Python file, use Copilot Chat:
@APEX help me implement a rate limiter
@TENSOR design ML pipeline for this task
@FORTRESS review security of this code
```

## Running Tests

### Unit Tests

```bash
cd backend
make test
# or:
go test -v ./...
```

### Integration Tests

```bash
cd backend
make test-integration
```

### Full Test Suite

```bash
cd backend
make test-all
```

### Python Test Suite

```bash
cd tests
python -m pytest supreme_master_suite/
```

## Troubleshooting Installation

### Port 8080 Already in Use

```bash
# Find process using port 8080
lsof -i :8080
# or on Windows:
netstat -ano | findstr :8080

# Kill the process or use a different port:
PORT=9090 go run ./cmd/server
```

### Build Fails with Go Version Error

```
Error: requires go1.21 or higher

# Solution: Update Go
# Visit https://golang.org/dl/ and download Go 1.21+
# Verify: go version
```

### Agents Not Loading

```
Warning: RegisterAllAgents returned error: ...

# Verify .github/agents directory exists:
ls ../.github/agents

# Check agent files are present:
ls ../.github/agents/*.agent.md | wc -l
# Should output: 40

# Check file permissions (Unix):
ls -la ../.github/agents/
```

### YAML Parsing Error

If you see YAML parsing errors:

```
Error: failed to parse YAML frontmatter: ...

# Check YAML format in .agent.md files:
# - Must start with ---
# - Must end with ---
# - Valid YAML syntax
# - Proper indentation (spaces, not tabs)
```

## Configuration

### Environment Variables

```bash
# Port to listen on (default: 8080)
PORT=9090 go run ./cmd/server

# Log level (debug, info, warn, error)
LOG_LEVEL=debug go run ./cmd/server

# Agent directory (for development)
AGENTS_DIR=/custom/path/.github/agents go run ./cmd/server

# OIDC Configuration (if using authentication)
OIDC_PROVIDER=https://github.com/login/oauth
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
```

### Configuration Files

The system uses configuration files in `backend/config/`:

- `config.yaml` - Main configuration
- `agents-manifest.yaml` - Agent registry (auto-generated)

## Next Steps

1. **Quick Start** - See [QUICK_START.md](QUICK_START.md)
2. **Agent Reference** - See [AGENT_LOADING_GUIDE.md](AGENT_LOADING_GUIDE.md)
3. **API Documentation** - See [API_REFERENCE.md](API_REFERENCE.md)
4. **Developer Guide** - See [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)

## Support

- **Documentation** - Check [docs/](../) directory
- **GitHub Issues** - Report problems on GitHub
- **GitHub Discussions** - Ask questions in Discussions
- **Contributing** - See [CONTRIBUTING.md](../CONTRIBUTING.md)

## Version Information

- **Current Version:** 2.0.0
- **Go Version:** 1.21+
- **Last Updated:** December 11, 2025
- **Status:** Production Ready

## License

MIT License - See [LICENSE](../LICENSE) for details.
