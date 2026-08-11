# Elite Agent Collective Backend

This is the HTTP backend service that powers the GitHub Copilot Extension for the Elite Agent Collective. It hosts logic for all 40 agents and exposes endpoints that Copilot can call for each agent.

## Features

- **40 Specialized Agents**: Full registry of all Elite Agent Collective agents
- **MNEMONIC Memory System**: Sub-linear experience retrieval with Bloom Filter (O(1)), LSH Index (O(1)), and HNSW Graph (O(log n))
- **ReMem-Elite Control Loop**: 5-phase memory-augmented execution (RETRIEVE → THINK → ACT → REFLECT → EVOLVE)
- **Experience-Based Learning**: Agents accumulate strategies, share knowledge across tiers, and self-improve at inference time
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

## Authority Levels

Every agent invocation is checked against the **authority ladder** before dispatch. A request
declares the level it wants with an optional header:

```
X-Agent-Level: L2
```

Accepted values are `L0`–`L4`, or the level's name (`perceive`, `converse`, `propose`, `act`,
`human-gated`), case-insensitive.

| Level | Name | Permits |
|-------|------|---------|
| `L0` | Perceive | Metadata only — state, logs, counters |
| `L1` | Converse | Plain-language answers from actual state |
| `L2` | Propose | Signed, structured, logged proposals — human-approved |
| `L3` | Act (scoped) | A narrow pre-authorised capability set; tools, not root |
| `L4` | Human-gated | Irreversibles — agent-prepared, human-performed |

**Ceilings are derived from each agent's existing tier**, so no per-agent configuration exists or
is needed:

| Tier | Ceiling | Agents |
|------|---------|--------|
| 4 | `L2` | `OMNISCIENT` — may propose, never act |
| 5 | `L3` | Domain specialists (`SENTRY`, `VERTEX`, `STREAM`, `ATLAS`, `FORGE`) |
| all others | `L2` | Tiers 1–3 and 6–8, **and any unrecognised tier** |

### A ceiling is not a grant

⚠️ **Reaching `L3` is necessary but not sufficient.** From `L3` upward a request must also
**name a capability**, and that capability must appear in the agent's enumerated grant table —
*"per-agent enumerated grants; nothing ambient."* A tier is a class and sets the ceiling; a grant
is a decision about one agent and says what it may do there.

**The grant table is currently EMPTY, and that is correct:** no capability is implemented and there
is no tool surface, so an entry would grant an ability that does not exist. **L3 is therefore
unreachable by construction** — every `L3` request is refused regardless of tier, and a test
fails if a grant ever appears, so granting power cannot ride along quietly with other work.

Capabilities narrow along a delegation chain exactly as ceilings do: **every** agent in the chain
must hold the capability, so a caller cannot borrow a callee's grant.

Behaviour:

- **No header means `L1`**, which is below every ceiling — so requests that worked before the
  ladder existed still work unchanged.
- A request **above** an agent's ceiling returns **`403`** with a reason. It is refused, never
  silently downgraded: a caller must be able to tell "I acted" from "I only spoke".
- An **unparseable** level returns **`400`**. It is not treated as the default; a caller asking for
  a level it cannot name is the caller to stop.
- Authority **narrows along a delegation chain, never widens**. When `OMNISCIENT` dispatches to a
  specialist, the effective ceiling is the minimum over the chain, so a tier-4 caller cannot reach
  `L3` by proxy through a tier-5 callee.
- Denials are counted in `eac_ladder_denials_total` and are **not** counted as dispatches. The
  check runs before `eac_requests_by_agent_total` increments, so a refusal is distinguishable from
  an invocation in the metrics.

```bash
# allowed — OMNISCIENT may propose
curl -X POST localhost:8080/agents/OMNISCIENT/invoke \
  -H 'Content-Type: application/json' -H 'X-Agent-Level: L2' \
  -d '{"messages":[{"role":"user","content":"outline an approach"}],"stream":false}'

# refused (403) — tier 4 is capped at L2
curl -X POST localhost:8080/agents/OMNISCIENT/invoke \
  -H 'Content-Type: application/json' -H 'X-Agent-Level: L3' \
  -d '{"messages":[{"role":"user","content":"act on this"}],"stream":false}'
```

The header is honoured on both `/agents/{codename}/invoke` and `/copilot`.

## Proposals (L2)

`L2 Propose` is defined by the ladder as *"signed, structured, Echo-logged proposals — human-approved."*
Three of those four hold here; the fourth is stated rather than implied.

```
POST /agents/{codename}/propose     -> 201 with a Proposal, or 422 if the agent did not produce one
GET  /proposals                     -> all proposals, newest first
GET  /proposals/{id}
POST /proposals/{id}/decision       -> {"decision":"approve"|"reject","note":"..."}
```

Enabled by `EAC_PROPOSAL_DIR`. Unset, every endpoint above returns **503** and nothing else changes.

**Structured, not prose.** The agent is asked for a JSON object and the response is parsed strictly.
If it returns prose, malformed JSON, or the instruction template echoed back, the request fails with
**422 and no proposal is stored** — there is deliberately no path that wraps prose in a
proposal-shaped envelope, because that would make L2 a label on an L1 answer.

**Commitment, not signature.** Each proposal carries a BLAKE2b-256 digest over its canonical content,
recorded in the audit chain at creation. Because the chain is append-only and hash-linked, content
cannot be altered afterwards without breaking it. That is **tamper-evidence, not authenticity**: it
proves the content is what was recorded, and says nothing about who produced it. Nothing is signed,
because there is no key-custody model in this service and a signature from an unmanaged key on the
serving host would be decorative.

The commitment excludes `status` and `decision`, so ruling on a proposal does not invalidate it. The
store verifies the commitment on **every read** and refuses to return altered content.

**`reversible_claimed` is the model's claim, not a fact.** §7 rates "agent error at machine speed" on
L3 being reversible; a bare `reversible: true` supplied by the model and checked by nobody would be a
claim carrying the authority of a measurement. The field name carries its provenance, as
`"authenticated": false` does on decisions. Authoritative reversibility will come from the capability a
step exercises, once L3 capabilities exist.

**Approval decides; it does not execute.** There is no tool surface in this service and L3 is not
implemented. An approved proposal is a decision on the record, not an action taken.

⚠️ **Approval is not currently authenticated.** Auth is enabled only when `OIDC_CLIENT_ID` is set. When
it is not, the stored decision records `"authenticated": false` and `"approver": "unauthenticated"` —
the record carries its own weakness, so a later reader can tell a verified approval from an unverified
one without knowing how the service was configured at the time.

A proposal is ruled on **once**; a second decision returns **409**.

## Audit Chain

When `EAC_ECHO_PATH` is set, every ladder decision — **allow and deny** — is appended to an
append-only, hash-chained log (BLAKE2b-256, length-prefixed fields, domain-separated, anchored by a
genesis record).

Records hold **hashes and metadata only**: agent codename, delegation chain, the three levels, and
the denied flag. **No prompt text, no message content, no keys** ever reach the log — which is what
makes it safe to leak, and is enforced by a test that puts a canary string in a request body and
asserts it appears in no byte of the chain.

- The chain is **verified in full at startup**. A log that does not verify **disables auditing**
  rather than being extended — appending onto a break would launder it behind valid-looking
  successors.
- Each append is fsynced and **read back** before the in-memory head advances; a write that is not
  durable, or that lands corrupt, is rolled back rather than chained onto.
- Audit failure is **fail-open below `L3`, fail-closed from `L3` up.**
  Below `L3` an agent produces words and no effect, so a lost record loses the evidence of a
  conversation — refusing to answer because a disk write failed would turn a logging problem into
  an outage for no integrity gain. From `L3` an agent **acts**, and a lost record means an action
  happened with no evidence that it did: *if it cannot be logged, it must not happen.*
- **There is no escape hatch, deliberately.** A flag permitting action without an audit record is the
  setting that gets turned on during an incident and never turned off. The consequence is stated
  plainly: **`L3` requires a working audit chain** — if `EAC_ECHO_PATH` is unset or the chain is
  broken, `L3` is unavailable. That is the rule, not a side effect of it.
- A fail-closed denial is itself unrecordable (the chain is what failed), so it is counted twice:
  `eac_echo_append_failures_total` **and** `eac_ladder_denials_total`. The pair is diagnostic —
  denials without append failures are ordinary ladder refusals; denials **with** them are the audit
  chain taking `L3` offline.
- Unset `EAC_ECHO_PATH` leaves auditing off and the request path unchanged.

## Observability

`GET /metrics` serves Prometheus text format.

| Metric | Type | Meaning |
|--------|------|---------|
| `eac_requests_total` | counter | HTTP requests served |
| `eac_errors_total` | counter | Responses with a 5xx status — **note: cannot see a request that dies when the connection closes mid-handler** |
| `eac_in_flight_requests` | gauge | Requests currently in flight |
| `eac_requests_by_agent_total{agent}` | counter | Agent dispatches by codename |
| `eac_routing_mode_total{mode}` | counter | Routing decisions (`explicit`/`semantic`/`crew`/`default`) |
| `eac_upstream_answers_total` | counter | Answers served by the upstream Copilot API |
| `eac_ollama_fallback_total` | counter | Answers served by the local model |
| `eac_model_timeouts_total` | counter | Model calls that exceeded their budget |
| `eac_ladder_denials_total{agent}` | counter | Requests refused by the authority ladder |
| `eac_echo_append_failures_total` | counter | Audit-chain appends that failed |

All counters **reset on restart**. Comparing a counter across a restart measures the restart, not
the change under test.

## Configuration

The server can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `BIND_ADDR` | `` | Listen address. Empty binds **all interfaces**; set it to pin the service to one. (A service reachable on a mesh interface but not loopback — or the reverse — is a recurring source of false "it is down" readings: verify the LISTENER, not the unit state.) |
| `LOG_LEVEL` | `info` | Logging level |
| `CORS_ALLOWED_ORIGINS` | `` | Comma-separated allowed origins |
| `OIDC_ISSUER` | `https://token.actions.githubusercontent.com` | OIDC issuer URL |
| `OIDC_CLIENT_ID` | `` | OIDC client ID (**enables authentication when set** — unset means the invoke route is unauthenticated) |
| `OIDC_CLIENT_SECRET` | `` | OIDC client secret — **secret, never commit** |
| `GITHUB_APP_ID` | `` | GitHub App ID used for webhook signature verification |
| `GITHUB_APP_PRIVATE_KEY` | `` | GitHub App private key — **secret, never commit** |
| `GITHUB_WEBHOOK_SECRET` | `` | Webhook signing secret — **secret, never commit** |
| `OLLAMA_URL` | `http://localhost:8000` | Model endpoint. Point at a gateway to move inference off the local host |
| `OLLAMA_MODEL` | `ling-mini-2.0` | Model name. ⚠️ A gateway may **route by model name**, so this and `OLLAMA_URL` must change together |
| `COPILOT_UPSTREAM_URL` | `` | Upstream Copilot API; unset disables upstream |
| `MEMORY_DATA_DIR` | `data/memories` | MNEMONIC store location |
| `CREWS_CONFIG` | `` | Path to the crew definitions file |
| `AGENTMEM_MCP_URL` | `` | agentmem MCP endpoint |
| `EAC_TASKD_URL` | `` | task daemon proxied under `/tasks` |
| `EMBED_URL` | `` | Embedding endpoint used by retrieval |
| `EMBED_MODEL` | `nomic-embed-text` | Embedding model |
| `EAC_RETRIEVER` | `` | Set to `1` to ground answers in the sigma-index corpus. Fail-open: an index outage degrades answers, never breaks invocations |
| `EAC_RETRIEVER_K` | `5` | Snippets retrieved per request |
| `EAC_RETRIEVER_INJECT` | `point-of-need` | `point-of-need` or `system` |
| `EAC_ECHO_PATH` | `` | Audit-chain log path. Unset disables auditing. Must be writable by the service account |
| `EAC_ECHO_IDENTITY` | `sigma-agentos` | Identity bound into the chain's genesis record |
| `EAC_PROPOSAL_DIR` | `` | L2 proposal store directory. Unset disables the propose/proposals endpoints. Must be writable by the service account |

### Memory System Configuration

The MNEMONIC memory system can be configured via `pkg/models/memory.go`:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `Enabled` | `true` | Enable/disable memory system |
| `EmbeddingDimension` | `384` | Size of embedding vectors |
| `MaxExperiencesPerAgent` | `1000` | Max experiences stored per agent |
| `MinFitnessThreshold` | `0.3` | Minimum fitness score for retrieval |
| `BreakthroughThreshold` | `0.9` | Fitness threshold for collective promotion |
| `LSHNumTables` | `10` | Number of LSH hash tables |
| `LSHNumHashFuncs` | `12` | Hash functions per table |
| `HNSWMaxConnections` | `16` | HNSW graph max connections (M parameter) |
| `HNSWEfConstruction` | `200` | HNSW construction quality parameter |
| `HNSWEfSearch` | `100` | HNSW search quality parameter |

**Tuning Guidelines:**
- Increase `LSHNumTables` for higher recall (more memory usage)
- Increase `HNSWEfConstruction` for better graph quality (slower indexing)
- Increase `HNSWEfSearch` for better recall (slower queries)
- Adjust `MinFitnessThreshold` to filter low-quality experiences

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
│   ├── copilot/
│   │   ├── request.go              # Copilot request parsing
│   │   └── response.go             # Copilot response formatting
│   └── memory/                     # MNEMONIC Memory System
│       ├── experience.go           # ExperienceTuple data structures, query contexts
│       ├── remem_loop.go           # ReMem-Elite control loop orchestration
│       ├── sublinear_retriever.go  # Sub-linear retrieval (Bloom, LSH, HNSW)
│       ├── errors.go               # Memory-specific error types
│       └── sublinear_retriever_test.go  # Comprehensive tests and benchmarks
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

### Memory System Components

The `internal/memory/` directory implements the MNEMONIC system:

| File | Purpose | Key Components |
|------|---------|----------------|
| **experience.go** | Core data structures | ExperienceTuple, QueryContext, RetrievalResult, Breakthrough, MemoryStats |
| **remem_loop.go** | ReMem control loop | ReMemController, ContextConstructor, MemoryUpdater, OutcomeEvaluator |
| **sublinear_retriever.go** | Sub-linear retrieval | BloomFilter (O(1)), LSHIndex (O(1)), HNSWGraph (O(log n)), SubLinearRetriever |
| **errors.go** | Error handling | Memory-specific error types and constants |
| **sublinear_retriever_test.go** | Testing & benchmarks | Unit tests, integration tests, performance benchmarks |

**Performance Characteristics:**
- Exact task signature matching: O(1) via Bloom Filter
- Approximate nearest neighbor: O(1) expected via LSH Index  
- Semantic similarity search: O(log n) via HNSW Graph
- Optimized for 1M+ experiences with minimal memory overhead

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
