# Quick Start Guide

Get the Elite Agent Collective up and running in 5 minutes.

## 1. Prerequisites (1 min)

Ensure you have:

- **Go 1.21+** installed (`go version`)
- **Git** installed (`git --version`)
- **VS Code** (optional, for best experience)

## 2. Clone & Setup (2 min)

```bash
# Clone the repository
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git
cd elite-agent-collective

# Navigate to backend
cd backend

# Install dependencies
go mod download
```

## 3. Start the Server (1 min)

```bash
# Using Makefile (recommended)
make run

# Or manually
go run ./cmd/server
```

**You should see:**

```
2025/12/11 10:00:00 Starting Elite Agent Collective backend server
2025/12/11 10:00:00 Loaded 40 agents from .../agents
2025/12/11 10:00:00 Server listening on :8080
```

## 4. Verify Installation (1 min)

Open a new terminal and test the API:

```bash
# Health check
curl http://localhost:8080/health

# List all agents
curl http://localhost:8080/agents | jq .

# Get a specific agent
curl http://localhost:8080/agents/APEX | jq .

# Invoke an agent
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "APEX",
    "messages": [{"role": "user", "content": "Help me design a rate limiter"}]
  }'
```

## Using with GitHub Copilot

### In VS Code Chat

```
@APEX help me implement a distributed cache
@CIPHER review this code for security issues
@ARCHITECT design a microservices architecture
@TENSOR design an ML pipeline
```

### Invoke Any of 40 Agents

All agent codenames are available:

**Tier 1 - Foundational:**

```
@APEX @CIPHER @ARCHITECT @AXIOM @VELOCITY
```

**Tier 2 - Specialists:**

```
@QUANTUM @TENSOR @FORTRESS @NEURAL @CRYPTO @FLUX @PRISM @SYNAPSE @CORE @HELIX @VANGUARD @ECLIPSE
```

**Tier 3 - Innovators:**

```
@NEXUS @GENESIS @OMNISCIENT
```

**Tier 5-8 - Domain & Enterprise:**

```
@ATLAS @FORGE @SENTRY @VERTEX @STREAM @PHOTON @LATTICE @MORPH
@PHANTOM @ORBIT @CANVAS @LINGUA @SCRIBE @MENTOR @BRIDGE
@AEGIS @LEDGER @PULSE @ARBITER @ORACLE
```

## Common Tasks

### Add a New Agent

1. Create file: `.github/agents/NEWAGENT.agent.md`
2. Copy from APEX.agent.md as template
3. Update metadata (name, codename, id, description)
4. Update content (philosophy, capabilities, examples)
5. Restart server - no code changes needed!

### Invoke Agent via API

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "CIPHER",
    "messages": [
      {
        "role": "user",
        "content": "Design a secure authentication system"
      }
    ]
  }'
```

### Run Tests

```bash
cd backend
make test          # Unit tests
make test-integration  # Integration tests
make test-all      # Full suite
```

## Directory Structure

```
elite-agent-collective/
├── .github/agents/
│   ├── APEX.agent.md
│   ├── CIPHER.agent.md
│   └── ... (40 agents total)
├── backend/
│   ├── cmd/server/      (main application)
│   ├── internal/agents/ (agent registry & handlers)
│   ├── go.mod          (dependencies)
│   └── Makefile        (build commands)
├── docs/               (documentation)
├── tests/              (test suite)
└── README.md           (overview)
```

## Memory System (MNEMONIC)

Each agent learns from past tasks and improves over time:

- **Experience Storage** - Agents store successful strategies
- **Cross-Tier Learning** - Agents share knowledge within their tier
- **Breakthrough Propagation** - Exceptional solutions spread to all agents
- **Fitness Evolution** - Strategies improve based on success metrics

This means agents get smarter with each use!

## Next Steps

1. **Explore Agents** - Try different agents in VS Code Chat
2. **Review Documentation** - See [docs/](../docs/) for detailed guides
3. **Run Tests** - Execute test suite to validate setup
4. **Integrate with CI/CD** - Deploy to production (Docker, Kubernetes)
5. **Contribute** - See [CONTRIBUTING.md](../CONTRIBUTING.md)

## Troubleshooting

**Server won't start?**

```bash
# Check Go version
go version  # Should be 1.21+

# Check dependencies
go mod tidy

# Check logs
go run ./cmd/server  # See error messages
```

**Port 8080 in use?**

```bash
# Use different port
PORT=9090 go run ./cmd/server
```

**Agents not loading?**

```bash
# Verify agent files exist
ls ../.github/agents/*.agent.md | wc -l  # Should be 40

# Check logs for directory resolution
go run ./cmd/server  # Look for "Loaded X agents from..."
```

## Resources

| Resource          | Link                                             |
| ----------------- | ------------------------------------------------ |
| Full Installation | [INSTALLATION.md](INSTALLATION.md)               |
| Agent Loading     | [AGENT_LOADING_GUIDE.md](AGENT_LOADING_GUIDE.md) |
| API Reference     | [API_REFERENCE.md](API_REFERENCE.md)             |
| Developer Guide   | [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)         |
| Troubleshooting   | [TROUBLESHOOTING.md](TROUBLESHOOTING.md)         |

## Support

- **Issues** → GitHub Issues
- **Questions** → GitHub Discussions
- **Contributing** → See [CONTRIBUTING.md](../CONTRIBUTING.md)

---

**Congrats! You're ready to use the Elite Agent Collective! 🚀**

Try this right now:

```
@APEX design a data pipeline for my project
```
