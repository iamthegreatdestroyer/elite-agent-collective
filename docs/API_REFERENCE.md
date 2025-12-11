# REST API Reference

Complete reference for Elite Agent Collective HTTP API.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently the API is **unauthenticated** in development mode. In production, enable OIDC:

```bash
OIDC_PROVIDER=https://github.com/login/oauth go run ./cmd/server
```

## Endpoints

### 1. Health Check

**Endpoint:** `GET /health`

**Description:** Check if server is running and agents are loaded

**Response:**

```json
{
  "status": "healthy",
  "agents_loaded": 40,
  "timestamp": "2025-12-11T10:00:00Z"
}
```

**Example:**

```bash
curl http://localhost:8080/health
```

---

### 2. List All Agents

**Endpoint:** `GET /agents`

**Description:** Get list of all available agents with metadata

**Query Parameters:**

- `tier` (optional) - Filter by tier (1-8)
- `category` (optional) - Filter by category

**Response:**

```json
{
  "agents": [
    {
      "id": "01",
      "codename": "APEX",
      "name": "Elite Computer Science Engineering",
      "tier": 1,
      "specialty": "Master-level software engineering, system design, and computational problem-solving",
      "philosophy": "Every problem has an elegant solution waiting to be discovered.",
      "directives": [
        "Production-grade, enterprise-quality code generation",
        "Data structures & algorithms at the deepest level",
        "System design & distributed systems architecture"
      ],
      "keywords": ["engineering", "design", "algorithms"],
      "examples": [
        "@APEX implement a rate limiter with sliding window",
        "@APEX design distributed cache system"
      ],
      "collaborators": ["ARCHITECT", "VELOCITY", "ECLIPSE"],
      "category": "Foundational"
    }
    // ... 39 more agents
  ],
  "total": 40,
  "timestamp": "2025-12-11T10:00:00Z"
}
```

**Examples:**

```bash
# Get all agents
curl http://localhost:8080/agents

# Filter by tier
curl http://localhost:8080/agents?tier=1

# Filter by category
curl http://localhost:8080/agents?category=Foundational

# Pretty print JSON
curl http://localhost:8080/agents | jq .
```

---

### 3. Get Specific Agent

**Endpoint:** `GET /agents/{codename}`

**Description:** Get detailed information about a specific agent

**Path Parameters:**

- `codename` (required) - Agent codename (e.g., APEX, CIPHER)

**Response:**

```json
{
  "id": "01",
  "codename": "APEX",
  "name": "Elite Computer Science Engineering",
  "tier": 1,
  "specialty": "Master-level software engineering, system design, and computational problem-solving",
  "philosophy": "Every problem has an elegant solution waiting to be discovered.",
  "directives": [
    "Production-grade, enterprise-quality code generation",
    "Data structures & algorithms at the deepest level",
    "System design & distributed systems architecture"
  ],
  "keywords": ["engineering", "design", "algorithms"],
  "examples": [
    "@APEX implement a rate limiter with sliding window",
    "@APEX design distributed cache system",
    "@APEX analyze this algorithm complexity",
    "@APEX review code for design patterns"
  ],
  "collaborators": ["ARCHITECT", "VELOCITY", "ECLIPSE"],
  "category": "Foundational"
}
```

**Examples:**

```bash
# Get APEX agent
curl http://localhost:8080/agents/APEX

# Get CIPHER agent
curl http://localhost:8080/agents/CIPHER

# With pretty printing
curl http://localhost:8080/agents/TENSOR | jq .
```

**Error Response (404):**

```json
{
  "error": "agent not found",
  "codename": "INVALID"
}
```

---

### 4. Invoke Agent

**Endpoint:** `POST /agent`

**Description:** Send request to an agent and get response

**Request Headers:**

```
Content-Type: application/json
Authorization: Bearer <token> (if authentication enabled)
```

**Request Body:**

```json
{
  "agent": "APEX",
  "messages": [
    {
      "role": "user",
      "content": "Help me implement a rate limiter with token bucket algorithm"
    }
  ],
  "context": {
    "language": "Python",
    "framework": "FastAPI",
    "use_case": "API rate limiting"
  }
}
```

**Parameters:**

- `agent` (required) - Agent codename
- `messages` (required) - Array of message objects
  - `role` (required) - "user" or "assistant"
  - `content` (required) - Message text
- `context` (optional) - Additional context for the agent

**Response:**

```json
{
  "agent": "APEX",
  "status": "success",
  "response": {
    "role": "assistant",
    "content": "[Agent response with implementation and explanation]"
  },
  "metadata": {
    "processing_time_ms": 1250,
    "tokens_used": 1842,
    "model": "claude-sonnet-4.5"
  }
}
```

**Examples:**

```bash
# Simple invocation
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "APEX",
    "messages": [
      {
        "role": "user",
        "content": "Design a distributed rate limiter"
      }
    ]
  }'

# With context
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "CIPHER",
    "messages": [
      {
        "role": "user",
        "content": "Review JWT implementation"
      }
    ],
    "context": {
      "language": "Python",
      "framework": "FastAPI",
      "security_level": "high"
    }
  }'

# Multi-turn conversation
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "ARCHITECT",
    "messages": [
      {
        "role": "user",
        "content": "Design microservices architecture for e-commerce"
      },
      {
        "role": "assistant",
        "content": "[Previous response from agent]"
      },
      {
        "role": "user",
        "content": "How would we handle payment processing?"
      }
    ]
  }'
```

**Error Response (Agent Not Found):**

```json
{
  "error": "agent not found",
  "agent": "INVALID_AGENT",
  "status": 404
}
```

**Error Response (Invalid Request):**

```json
{
  "error": "invalid request",
  "details": "agent field is required",
  "status": 400
}
```

---

## Status Codes

| Code | Meaning                          |
| ---- | -------------------------------- |
| 200  | Success                          |
| 400  | Bad Request (invalid parameters) |
| 404  | Not Found (agent not found)      |
| 429  | Too Many Requests (rate limited) |
| 500  | Internal Server Error            |
| 503  | Service Unavailable              |

---

## Rate Limiting

API rate limits are applied per agent endpoint:

- **Free Tier:** 100 requests/hour per agent
- **Premium Tier:** 1000 requests/hour per agent
- **Header:** `X-RateLimit-Remaining` shows remaining requests

**Rate Limit Response:**

```json
{
  "error": "rate limit exceeded",
  "retry_after_seconds": 60,
  "limit": 100,
  "period": "hour"
}
```

---

## Agent Codenames Reference

### Tier 1 - Foundational

- `APEX` - Elite Computer Science Engineering
- `CIPHER` - Advanced Cryptography & Security
- `ARCHITECT` - Systems Architecture & Design Patterns
- `AXIOM` - Pure Mathematics & Formal Proofs
- `VELOCITY` - Performance Optimization & Sub-Linear Algorithms

### Tier 2 - Specialists

- `QUANTUM` - Quantum Mechanics & Quantum Computing
- `TENSOR` - Machine Learning & Deep Neural Networks
- `FORTRESS` - Defensive Security & Penetration Testing
- `NEURAL` - Cognitive Computing & AGI Research
- `CRYPTO` - Blockchain & Distributed Systems
- `FLUX` - DevOps & Infrastructure Automation
- `PRISM` - Data Science & Statistical Analysis
- `SYNAPSE` - Integration Engineering & API Design
- `CORE` - Low-Level Systems & Compiler Design
- `HELIX` - Bioinformatics & Computational Biology
- `VANGUARD` - Research Analysis & Literature Synthesis
- `ECLIPSE` - Testing, Verification & Formal Methods

### Tier 3 - Innovators

- `NEXUS` - Paradigm Synthesis & Cross-Domain Innovation
- `GENESIS` - Zero-to-One Innovation & Novel Discovery
- `OMNISCIENT` - Meta-Learning & Evolution Orchestrator

### Tier 5 - Domain Specialists

- `ATLAS` - Cloud Infrastructure & Multi-Cloud Architecture
- `FORGE` - Build Systems & Compilation Pipelines
- `SENTRY` - Observability, Logging & Monitoring
- `VERTEX` - Graph Databases & Network Analysis
- `STREAM` - Real-Time Data Processing & Event Streaming

### Tier 6 - Emerging Tech

- `PHOTON` - Edge Computing & IoT Systems
- `LATTICE` - Distributed Consensus & CRDT Systems
- `MORPH` - Code Migration & Legacy Modernization
- `PHANTOM` - Reverse Engineering & Binary Analysis
- `ORBIT` - Satellite & Embedded Systems Programming

### Tier 7 - Human-Centric

- `CANVAS` - UI/UX Design Systems & Accessibility
- `LINGUA` - Natural Language Processing & LLM Fine-Tuning
- `SCRIBE` - Technical Documentation & API Docs
- `MENTOR` - Code Review & Developer Education
- `BRIDGE` - Cross-Platform & Mobile Development

### Tier 8 - Enterprise

- `AEGIS` - Compliance, GDPR & SOC2 Automation
- `LEDGER` - Financial Systems & Fintech Engineering
- `PULSE` - Healthcare IT & HIPAA Compliance
- `ARBITER` - Conflict Resolution & Merge Strategies
- `ORACLE` - Predictive Analytics & Business Intelligence

---

## Examples by Use Case

### Code Review

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "ECLIPSE",
    "messages": [
      {
        "role": "user",
        "content": "Review this code for test coverage: [code snippet]"
      }
    ]
  }'
```

### Security Analysis

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "FORTRESS",
    "messages": [
      {
        "role": "user",
        "content": "Perform security audit on this API implementation"
      }
    ]
  }'
```

### Architecture Design

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "ARCHITECT",
    "messages": [
      {
        "role": "user",
        "content": "Design a microservices architecture for a SaaS platform"
      }
    ]
  }'
```

### ML/AI Development

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "TENSOR",
    "messages": [
      {
        "role": "user",
        "content": "Design a CNN architecture for image classification"
      }
    ]
  }'
```

---

## SDKs & Client Libraries

### Python

```python
import requests

response = requests.post(
    'http://localhost:8080/agent',
    json={
        'agent': 'APEX',
        'messages': [
            {
                'role': 'user',
                'content': 'Help me design a system'
            }
        ]
    }
)

print(response.json())
```

### JavaScript/Node.js

```javascript
const response = await fetch("http://localhost:8080/agent", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    agent: "APEX",
    messages: [
      {
        role: "user",
        content: "Help me design a system",
      },
    ],
  }),
});

console.log(await response.json());
```

### cURL (Bash)

```bash
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "APEX",
    "messages": [{"role": "user", "content": "Help me design a system"}]
  }' | jq .
```

---

## Webhooks

GitHub Copilot Extension can route requests as webhooks:

**Request Header:** `X-GitHub-Delivery`
**Signature Verification:** Use `X-Hub-Signature-256` header

---

## Changelog

### Version 2.0.0

- ✅ Dynamic agent loading from `.github/agents/` directory
- ✅ Extended Agent model with additional metadata fields
- ✅ Smart directory resolution with multiple fallback paths
- ✅ Graceful error handling and fallback to hardcoded definitions
- ✅ Full backward compatibility maintained

### Version 1.0.0

- Initial release with 40 hardcoded agents

---

## Support

- **Documentation** - See [docs/](../) directory
- **Issues** - Report problems on GitHub
- **Questions** - Ask in GitHub Discussions
