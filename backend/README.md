# Elite Agent Collective Backend

This is the HTTP backend service that powers the GitHub Copilot Extension for the Elite Agent Collective. It hosts logic for all 40 agents and exposes endpoints that Copilot can call for each agent.

## Features

- **40 Specialized Agents**: Full registry of all Elite Agent Collective agents
- **RESTful API**: Clean API design with proper error handling
- **GitHub Copilot Integration**: Compatible with GitHub Copilot Extension specifications
- **OIDC Authentication**: Stub implementation ready for OIDC integration
- **Graceful Shutdown**: Proper signal handling for clean shutdown
- **Health Checks**: Built-in health check endpoint for monitoring
- **Docker Support**: Containerized deployment with Docker and docker-compose
- **Logging**: Request logging middleware for debugging

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)

### Running Locally

```bash
# Install dependencies
make deps

# Run the server
make run
```

The server will start on `http://localhost:8080`.

### Using Docker

```bash
# Build Docker image
make docker

# Run Docker container
make docker-run
```

Or use docker-compose:

```bash
make compose-up
```

## API Endpoints

### Health Check

```
GET /health
```

Returns the server health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "service": "elite-agent-collective",
  "version": "1.0.0"
}
```

### List All Agents

```
GET /agents
```

Returns a list of all registered agents.

**Response:**
```json
[
  {
    "id": "01",
    "codename": "APEX",
    "tier": 1,
    "specialty": "Elite Computer Science Engineering",
    "philosophy": "Every problem has an elegant solution waiting to be discovered.",
    "directives": [...]
  },
  ...
]
```

### Get Agent Info

```
GET /agents/{codename}
```

Returns information about a specific agent.

**Parameters:**
- `codename`: The agent's codename (e.g., `APEX`, `CIPHER`, etc.)

**Response:**
```json
{
  "id": "01",
  "codename": "APEX",
  "tier": 1,
  "specialty": "Elite Computer Science Engineering",
  "philosophy": "Every problem has an elegant solution waiting to be discovered.",
  "directives": [
    "Deliver production-grade, enterprise-quality code",
    "Apply computer science fundamentals at the deepest level",
    ...
  ]
}
```

### Invoke Agent

```
POST /agents/{codename}/invoke
```

Invokes a specific agent with a Copilot request.

**Parameters:**
- `codename`: The agent's codename

**Request Body:**
```json
{
  "messages": [
    {"role": "user", "content": "Help me optimize this algorithm"}
  ],
  "model": "gpt-4",
  "stream": false
}
```

**Response:**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "As APEX, the Elite Computer Science Engineering Specialist..."
      },
      "finish_reason": "stop"
    }
  ]
}
```

### Copilot Webhook

```
POST /copilot
```

Main Copilot webhook endpoint. Automatically routes to the appropriate agent based on the message content (e.g., `@APEX help me`).

**Request Body:**
```json
{
  "messages": [
    {"role": "user", "content": "@APEX help me with this code"}
  ],
  "model": "gpt-4",
  "stream": false
}
```

## Configuration

The server can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `info` | Logging level |
| `OIDC_ISSUER` | `https://token.actions.githubusercontent.com` | OIDC issuer URL |
| `OIDC_CLIENT_ID` | `` | OIDC client ID (enables authentication when set) |
| `OIDC_CLIENT_SECRET` | `` | OIDC client secret |

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── agents/
│   │   ├── registry.go             # Agent registration and lookup
│   │   ├── handler.go              # HTTP handlers for agent endpoints
│   │   ├── agents.go               # Agent definitions and registration
│   │   └── handlers/
│   │       ├── base.go             # Base agent implementation
│   │       └── apex.go             # APEX agent implementation
│   ├── auth/
│   │   ├── oidc.go                 # OIDC authentication
│   │   └── middleware.go           # Auth middleware
│   ├── config/
│   │   └── config.go               # Configuration management
│   └── copilot/
│       ├── request.go              # Copilot request parsing
│       └── response.go             # Copilot response formatting
├── pkg/
│   └── models/
│       └── agent.go                # Agent data models
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Agent Registry

All 40 agents are registered at startup:

### Tier 1: Foundational Agents
- **APEX** (01) - Elite Computer Science Engineering
- **CIPHER** (02) - Advanced Cryptography & Security
- **ARCHITECT** (03) - Systems Architecture & Design Patterns
- **AXIOM** (04) - Pure Mathematics & Formal Proofs
- **VELOCITY** (05) - Performance Optimization & Sub-Linear Algorithms

### Tier 2: Specialist Agents
- **QUANTUM** (06) - Quantum Mechanics & Quantum Computing
- **TENSOR** (07) - Machine Learning & Deep Neural Networks
- **FORTRESS** (08) - Defensive Security & Penetration Testing
- **NEURAL** (09) - Cognitive Computing & AGI Research
- **CRYPTO** (10) - Blockchain & Distributed Systems
- **FLUX** (11) - DevOps & Infrastructure Automation
- **PRISM** (12) - Data Science & Statistical Analysis
- **SYNAPSE** (13) - Integration Engineering & API Design
- **CORE** (14) - Low-Level Systems & Compiler Design
- **HELIX** (15) - Bioinformatics & Computational Biology
- **VANGUARD** (16) - Research Analysis & Literature Synthesis
- **ECLIPSE** (17) - Testing, Verification & Formal Methods

### Tier 3: Innovator Agents
- **NEXUS** (18) - Paradigm Synthesis & Cross-Domain Innovation
- **GENESIS** (19) - Zero-to-One Innovation & Novel Discovery

### Tier 4: Meta Agents
- **OMNISCIENT** (20) - Meta-Learning Trainer & Evolution Orchestrator

### Tier 5: Domain Specialists
- **ATLAS** (21) - Cloud Infrastructure & Multi-Cloud Architecture
- **FORGE** (22) - Build Systems & Compilation Pipelines
- **SENTRY** (23) - Observability, Logging & Monitoring
- **VERTEX** (24) - Graph Databases & Network Analysis
- **STREAM** (25) - Real-Time Data Processing & Event Streaming

### Tier 6: Emerging Tech Specialists
- **PHOTON** (26) - Edge Computing & IoT Systems
- **LATTICE** (27) - Distributed Consensus & CRDT Systems
- **MORPH** (28) - Code Migration & Legacy Modernization
- **PHANTOM** (29) - Reverse Engineering & Binary Analysis
- **ORBIT** (30) - Satellite & Embedded Systems Programming

### Tier 7: Human-Centric Specialists
- **CANVAS** (31) - UI/UX Design Systems & Accessibility
- **LINGUA** (32) - Natural Language Processing & LLM Fine-Tuning
- **SCRIBE** (33) - Technical Documentation & API Docs
- **MENTOR** (34) - Code Review & Developer Education
- **BRIDGE** (35) - Cross-Platform & Mobile Development

### Tier 8: Enterprise & Compliance Specialists
- **AEGIS** (36) - Compliance, GDPR & SOC2 Automation
- **LEDGER** (37) - Financial Systems & Fintech Engineering
- **PULSE** (38) - Healthcare IT & HIPAA Compliance
- **ARBITER** (39) - Conflict Resolution & Merge Strategies
- **ORACLE** (40) - Predictive Analytics & Forecasting Systems

## Development

### Running Tests

```bash
make test
```

### Running Tests with Coverage

```bash
make test-coverage
```

### Formatting Code

```bash
make fmt
```

### Linting

```bash
make lint
```

## License

MIT License - see the [LICENSE](../LICENSE) file for details.
