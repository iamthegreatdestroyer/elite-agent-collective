# System Architecture

Detailed architectural overview of the Elite Agent Collective backend system.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      GitHub Copilot Chat                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  User: @APEX implement a rate limiter                            │
│                           │                                      │
│                           ▼                                      │
├─────────────────────────────────────────────────────────────────┤
│                  GitHub Copilot Extension API                   │
│         HTTP POST request to /agent endpoint                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│    ┌──────────────────────────────────────────────────────┐     │
│    │   Elite Agent Collective Backend (Go HTTP Server)    │     │
│    │                  Port 8080                           │     │
│    │                                                      │     │
│    │  ┌────────────────────────────────────────────────┐ │     │
│    │  │           Authentication Layer                │ │     │
│    │  │  (OIDC validation, signature verification)    │ │     │
│    │  └─────────────────────┬──────────────────────────┘ │     │
│    │                        │                            │     │
│    │  ┌────────────────────▼──────────────────────────┐ │     │
│    │  │       HTTP Router & Handler                  │ │     │
│    │  │  • GET /health                               │ │     │
│    │  │  • GET /agents                               │ │     │
│    │  │  • GET /agents/{codename}                    │ │     │
│    │  │  • POST /agent                               │ │     │
│    │  └─────────────────────┬──────────────────────────┘ │     │
│    │                        │                            │     │
│    │  ┌────────────────────▼──────────────────────────┐ │     │
│    │  │       Agent Registry & Loader                │ │     │
│    │  │  • 40 Distributed Agent Files                │ │     │
│    │  │  • .github/agents/*.agent.md                 │ │     │
│    │  │  • Dynamic loading at startup                │ │     │
│    │  │  • Thread-safe in-memory registry            │ │     │
│    │  └─────────────────────┬──────────────────────────┘ │     │
│    │                        │                            │     │
│    │  ┌────────────────────▼──────────────────────────┐ │     │
│    │  │       MNEMONIC Memory System                 │ │     │
│    │  │  • Experience Retrieval (O(1) to O(log n))   │ │     │
│    │  │  • ReMem Control Loop                        │ │     │
│    │  │  • Sub-Linear Algorithms                     │ │     │
│    │  │  • Experience Storage & Evolution            │ │     │
│    │  └─────────────────────┬──────────────────────────┘ │     │
│    │                        │                            │     │
│    │  ┌────────────────────▼──────────────────────────┐ │     │
│    │  │    Agent Handler & Response Builder          │ │     │
│    │  │  • Route to specific agent                    │ │     │
│    │  │  • Process agent request                      │ │     │
│    │  │  • Construct response JSON                    │ │     │
│    │  └──────────────────────────────────────────────┘ │     │
│    │                                                      │     │
│    └──────────────────────────────────────────────────────┘     │
│                                                                   │
│    ┌──────────────────────────────────────────────────────┐     │
│    │     Persistent Storage Layer (Optional)              │     │
│    │  • PostgreSQL: Agent usage analytics                │     │
│    │  • Redis: Request caching                           │     │
│    │  • File System: Agent definitions & config          │     │
│    │  • Vector DB: Experience embeddings                 │     │
│    └──────────────────────────────────────────────────────┘     │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Request Processing Pipeline

Detailed 8-step request processing flow:

```
┌────────────────────────────────────────────────────────────────┐
│ 1. Incoming HTTP Request                                       │
│    POST /agent                                                 │
│    {                                                           │
│      "agent": "APEX",                                         │
│      "messages": [{"role": "user", "content": "..."}]        │
│    }                                                           │
└──────────────────────────┬───────────────────────────────────┘
                           │
        ▼──────────────────────────────────────▼
┌────────────────────────────────────────────────────────────────┐
│ 2. Authentication Layer                                        │
│    • Extract OIDC token from Authorization header             │
│    • Verify GitHub webhook signature (X-Hub-Signature-256)    │
│    • Validate token expiration                                │
│    • On failure: Return 401 Unauthorized                      │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 3. Request Parsing & Validation                                │
│    • Parse JSON request body                                  │
│    • Extract agent codename (e.g., "APEX")                   │
│    • Validate required fields (agent, messages)               │
│    • Validate message format                                  │
│    • On error: Return 400 Bad Request                         │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 4. Agent Registry Lookup                                       │
│    • Look up agent in thread-safe registry                    │
│    • Retrieve agent metadata (name, tier, capabilities, etc.) │
│    • Check if agent is available                              │
│    • On not found: Return 404 Agent Not Found                 │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 5. MNEMONIC Memory Retrieval                                   │
│    • Query experience database with request context           │
│    • Retrieve relevant past solutions (O(1) to O(log n))      │
│    • Rank by relevance and fitness score                      │
│    • Augment context with learned strategies                  │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 6. Agent Invocation                                            │
│    • Call agent's Handle method                               │
│    • Pass request + retrieved memory context                  │
│    • Agent processes request (may call LLM, external APIs)    │
│    • Return structured response                               │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 7. Memory Storage & Evolution                                  │
│    • Store request/response in experience database            │
│    • Compute fitness score based on quality                   │
│    • Check for breakthrough discoveries (>0.9 fitness)        │
│    • Promote to collective memory if breakthrough             │
│    • Update agent tier-specific knowledge base                │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ 8. Response Construction & Delivery                            │
│    • Build response JSON with:                                │
│      - agent name and codename                               │
│      - status (success, error, etc.)                         │
│      - response content                                      │
│      - metadata (timestamp, request_id, etc.)                │
│    • Set appropriate HTTP status code (200, 400, 404, etc.)  │
│    • Return to GitHub Copilot Extension                      │
└────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. HTTP Server Layer

```
┌──────────────────────────────────────────┐
│     HTTP Server (net/http)               │
│     Port 8080 (configurable)             │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │      Router (Default Mux)          │  │
│  │                                    │  │
│  │  GET  /health           → handler  │  │
│  │  GET  /agents           → handler  │  │
│  │  GET  /agents/{codename}→ handler  │  │
│  │  POST /agent            → handler  │  │
│  │                                    │  │
│  └────────────────────────────────────┘  │
│                                          │
└──────────────────────────────────────────┘
```

### 2. Agent Registry

```
┌──────────────────────────────────────────┐
│    Agent Registry (Thread-Safe)          │
│                                          │
│  ┌─────────────────────────────────┐    │
│  │  RWMutex Lock                   │    │
│  │                                 │    │
│  │  agents map[string]Agent        │    │
│  │  ├─ "APEX"    → Agent{}         │    │
│  │  ├─ "CIPHER"  → Agent{}         │    │
│  │  ├─ "..."     → Agent{}         │    │
│  │  └─ "VERTEX"  → Agent{}         │    │
│  │                                 │    │
│  │  handlers map[string]Handler    │    │
│  │  ├─ "APEX"    → BaseHandler     │    │
│  │  ├─ "CIPHER"  → BaseHandler     │    │
│  │  ├─ "..."     → BaseHandler     │    │
│  │  └─ "VERTEX"  → BaseHandler     │    │
│  └─────────────────────────────────┘    │
│                                          │
│  Methods:                                │
│  • Register(handler) error               │
│  • GetInfo(codename) (*Agent, error)     │
│  • ListAgents() []Agent                  │
│  • Get(codename) (Handler, error)        │
│                                          │
└──────────────────────────────────────────┘
```

### 3. Agent Loader

```
┌──────────────────────────────────────────────┐
│       Agent Loader (.agent.md Parser)        │
│                                              │
│  Input: .github/agents/*.agent.md            │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │  YAML Frontmatter Parser             │   │
│  │  ───────────────────────             │   │
│  │  Extracts:                           │   │
│  │  • name, description, codename       │   │
│  │  • tier, id, category                │   │
│  │  • keywords, philosophy              │   │
│  │  • capabilities, examples            │   │
│  │  • collaborators, integration_notes  │   │
│  └──────────────────────────────────────┘   │
│                      │                      │
│                      ▼                      │
│  ┌──────────────────────────────────────┐   │
│  │  Markdown Section Extractor          │   │
│  │  ────────────────────────            │   │
│  │  Extracts:                           │   │
│  │  • Philosophy (##)                   │   │
│  │  • Capabilities (- bullet points)    │   │
│  │  • Methodology (numbered steps)      │   │
│  │  • Invocation Examples (###)         │   │
│  │  • Core Directives (blockquote)      │   │
│  │  • Technical Stack                   │   │
│  │  • Integration notes                 │   │
│  └──────────────────────────────────────┘   │
│                      │                      │
│                      ▼                      │
│  ┌──────────────────────────────────────┐   │
│  │  Agent Struct Constructor            │   │
│  │  ────────────────────────            │   │
│  │  Creates:                            │   │
│  │  struct Agent {                      │   │
│  │    ID, Codename, Tier, Specialty     │   │
│  │    Philosophy, Directives (orig)     │   │
│  │    Name, Keywords, Examples          │   │
│  │    Collaborators, Category, Path     │   │
│  │  }                                   │   │
│  └──────────────────────────────────────┘   │
│                      │                      │
│                      ▼                      │
│  Output: []Agent (40 agents loaded)         │
│                                              │
└──────────────────────────────────────────────┘
```

### 4. MNEMONIC Memory System

```
┌────────────────────────────────────────────┐
│      MNEMONIC Memory System                 │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  ReMem Control Loop                │   │
│  │  (RETRIEVE→THINK→ACT→REFLECT→EVOLVE) │   │
│  │                                    │   │
│  │  Phase 1: RETRIEVE                 │   │
│  │  ├─ Bloom Filter (O(1))            │   │
│  │  ├─ LSH Index (O(1))               │   │
│  │  └─ HNSW Graph (O(log n))          │   │
│  │                                    │   │
│  │  Phase 2: THINK                    │   │
│  │  ├─ Augment context with memory    │   │
│  │  └─ Format learned strategies      │   │
│  │                                    │   │
│  │  Phase 3: ACT                      │   │
│  │  ├─ Execute agent with context     │   │
│  │  └─ Apply learned patterns         │   │
│  │                                    │   │
│  │  Phase 4: REFLECT                  │   │
│  │  ├─ Evaluate execution outcome     │   │
│  │  └─ Compute fitness score          │   │
│  │                                    │   │
│  │  Phase 5: EVOLVE                   │   │
│  │  ├─ Store new experience           │   │
│  │  ├─ Promote breakthroughs          │   │
│  │  └─ Propagate successful strategies│   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  Sub-Linear Retrieval              │   │
│  │  (13 data structures)              │   │
│  │                                    │   │
│  │  Core Retrieval:                   │   │
│  │  1. Bloom Filter (O(1))            │   │
│  │  2. LSH Index (O(1))               │   │
│  │  3. HNSW Graph (O(log n))          │   │
│  │                                    │   │
│  │  Advanced Structures:              │   │
│  │  4. Count-Min Sketch (O(1))        │   │
│  │  5. Cuckoo Filter (O(1))           │   │
│  │  6. Product Quantizer              │   │
│  │  7. MinHash + LSH (O(1))           │   │
│  │                                    │   │
│  │  Agent-Aware Structures:           │   │
│  │  8. AgentAffinityGraph (O(1))      │   │
│  │  9. TierResonanceFilter            │   │
│  │  10. SkillBloomCascade             │   │
│  │  11. TemporalDecaySketch (O(1))    │   │
│  │  12. CollaborativeAttentionIndex   │   │
│  │  13. EmergentInsightDetector (O(1))│   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  Experience Storage                │   │
│  │  ────────────────────              │   │
│  │  ExperienceTuple:                  │   │
│  │  • Input (task, context)           │   │
│  │  • Output (solution, code)         │   │
│  │  • Strategy (approach used)        │   │
│  │  • Embedding (semantic vector)     │   │
│  │  • FitnessScore (quality metric)   │   │
│  │  • Timestamp (recency)             │   │
│  │  • AgentID, TierID (routing)       │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
└────────────────────────────────────────────┘
```

### 5. Agent Handler

```
┌────────────────────────────────────────────┐
│      Agent Handler Interface                │
│                                            │
│  Interface AgentHandler {                  │
│    GetInfo() *Agent                        │
│    Handle(req *Request) (*Response, error) │
│  }                                         │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  Base Agent Handler                │   │
│  │  (Concrete Implementation)         │   │
│  │                                    │   │
│  │  GetInfo()                         │   │
│  │  └─ Returns agent metadata         │   │
│  │                                    │   │
│  │  Handle(request)                   │   │
│  │  ├─ Retrieve from memory           │   │
│  │  ├─ Prepare agent context          │   │
│  │  ├─ Call agent logic               │   │
│  │  ├─ Process response               │   │
│  │  ├─ Store in memory                │   │
│  │  └─ Return response                │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  Specialized Handlers (Optional)   │   │
│  │                                    │   │
│  │  • LLMAgentHandler (Claude, GPT)   │   │
│  │  • ToolAgentHandler (external)     │   │
│  │  • HybridAgentHandler (combined)   │   │
│  │  • CustomAgentHandler (user-impl)  │   │
│  │                                    │   │
│  └────────────────────────────────────┘   │
│                                            │
└────────────────────────────────────────────┘
```

## Data Flow Diagrams

### Agent Request Data Flow

```
Request:                 Response:
┌────────┐              ┌────────┐
│ Agent  │              │ Agent  │
│ APEX   │              │ APEX   │
└────┬───┘              └───┬────┘
     │                      │
     │                      │
     ▼                      │
┌────────────────┐         │
│ Messages       │         │
│ [{role, text}] │         │
└────┬───────────┘         │
     │                     │
     │                     │
     ▼                     │
┌──────────────────┐       │
│ Context (opt)    │       │
│ {language, ...}  │       │
└────┬─────────────┘       │
     │                     │
     │                     │
     ▼                     │
┌──────────────────────────────┐
│  Agent Handler Process        │
│                              │
│ 1. Retrieve from memory      │
│ 2. Augment context           │
│ 3. Execute agent logic       │
│ 4. Generate response         │
│ 5. Store in memory           │
│                              │
└──────────┬───────────────────┘
           │
           ▼
      ┌──────────────┐
      │ Status       │
      │ success|err  │
      └──────┬───────┘
             │
             ▼
      ┌──────────────┐
      │ Response     │
      │ content      │
      │ text/code    │
      └──────┬───────┘
             │
             ▼
      ┌──────────────┐
      │ Metadata     │
      │ timestamp    │
      │ request_id   │
      └──────┬───────┘
             │
             ▼────────────────────► Response to client
```

### Agent Loading Data Flow

```
File System:                Parser:                    Registry:
.github/agents/    ──────►  Agent Loader   ──────►  In-Memory
 • APEX.agent.md            • Parse YAML             Map
 • CIPHER.agent.md          • Extract markdown      {
 • CRYPTO.agent.md          • Build Agent struct     "APEX": Agent{...},
 • FLUX.agent.md                                     "CIPHER": Agent{...},
 • ... (40 total)           ┌──────────┐             "CRYPTO": Agent{...},
                            │ 40 Agents│             ...
                            └──────────┘           }
                                 │
                                 ▼
                            Validation
                            ├─ ID unique?
                            ├─ Codename valid?
                            └─ Tier correct?
                                 │
                                 ▼
                            Registration
                            ├─ Add to registry
                            └─ Create handlers
```

## Concurrency & Thread Safety

### Registry Thread Safety

```
┌──────────────────────────────────────────┐
│      Agent Registry                      │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  RWMutex                           │  │
│  │  (Read-Write Lock)                 │  │
│  │                                    │  │
│  │  Multiple Readers (concurrent):   │  │
│  │  ├─ GET /agents (read lock)       │  │
│  │  ├─ GET /agents/{id} (read lock)  │  │
│  │  └─ Handler lookup (read lock)    │  │
│  │                                    │  │
│  │  Single Writer (exclusive):       │  │
│  │  └─ Register (write lock)         │  │
│  │                                    │  │
│  └────────────────────────────────────┘  │
│                                          │
│  Pattern:                                │
│  • Lock held for minimum duration      │
│  • Copy data before releasing lock     │
│  • No nested lock acquisitions         │
│                                          │
└──────────────────────────────────────────┘
```

## Scalability Considerations

### Horizontal Scaling

```
┌──────────────────────────────────────────────────┐
│  Load Balancer (Nginx, HAProxy, AWS ALB)         │
└─────────┬───────────────┬───────────┬─────────────┘
          │               │           │
          ▼               ▼           ▼
    ┌─────────┐      ┌─────────┐  ┌─────────┐
    │Instance │      │Instance │  │Instance │
    │   #1    │      │   #2    │  │   #3    │
    │Port 8080│      │Port 8080│  │Port 8080│
    └────┬────┘      └────┬────┘  └────┬────┘
         │                │           │
         └────────┬───────┴───────────┘
                  │
         ┌────────▼────────┐
         │  Shared Cache   │
         │  (Redis)        │
         └────────────────┘
```

### Vertical Scaling

```
Performance Tuning:
┌─────────────────────────────────────┐
│ • Optimize agent loading (YAML)     │
│ • Cache agent metadata              │
│ • Use connection pooling            │
│ • Profile and optimize hot paths    │
│ • Monitor memory usage              │
│ • Batch LLM requests                │
│ • Implement request queuing         │
└─────────────────────────────────────┘
```

## Deployment Architecture

### Docker Containerization

```
┌─────────────────────────────────────────┐
│  Docker Image                           │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  Base Image: golang:1.21-alpine   │  │
│  │  ├─ Go runtime                    │  │
│  │  └─ Minimal dependencies          │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  Build Stage                      │  │
│  │  ├─ Copy source code              │  │
│  │  ├─ go mod download               │  │
│  │  └─ go build                      │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  Runtime Stage                    │  │
│  │  ├─ Binary artifact               │  │
│  │  ├─ Alpine base                   │  │
│  │  └─ Port 8080 exposed             │  │
│  └───────────────────────────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

### Kubernetes Deployment

```
┌──────────────────────────────────────────┐
│  Kubernetes Cluster                      │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  Deployment                        │  │
│  │  • 3 replicas                      │  │
│  │  • Rolling updates                 │  │
│  │  • Resource limits (CPU, memory)   │  │
│  └────────────────────────────────────┘  │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  Service                           │  │
│  │  • Internal DNS: api.default.svc   │  │
│  │  • Port 8080 exposed               │  │
│  │  • Load balancing                  │  │
│  └────────────────────────────────────┘  │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  ConfigMap                         │  │
│  │  • config.yaml                     │  │
│  │  • agents-manifest.yaml            │  │
│  └────────────────────────────────────┘  │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  Secret                            │  │
│  │  • OIDC_CLIENT_SECRET              │  │
│  │  • GITHUB_WEBHOOK_SECRET           │  │
│  └────────────────────────────────────┘  │
│                                          │
└──────────────────────────────────────────┘
```

## Integration Points

### GitHub Copilot Integration

```
┌──────────────────┐
│ GitHub Copilot   │
│ Extension API    │
└────────┬─────────┘
         │
         │ HTTP POST /agent
         │ (with request body)
         │
         ▼
┌────────────────────────────┐
│ Elite Agent Collective     │
│ Backend (Port 8080)        │
└────────┬───────────────────┘
         │
         │ Structured Response
         │ (agent, status, content, metadata)
         │
         ▼
┌──────────────────┐
│ GitHub Copilot   │
│ Chat Interface   │
│                  │
│ Response shown   │
│ to user          │
└──────────────────┘
```

### External LLM Services

```
Elite Agent Collective
         │
         ├─→ Anthropic (Claude)
         │   • Fast responses
         │   • High quality
         │
         ├─→ OpenAI (GPT-4)
         │   • Fallback option
         │   • Specialized tasks
         │
         └─→ Local Models (Optional)
             • Self-hosted
             • Privacy-focused
             • Resource intensive
```

## Security Architecture

```
┌────────────────────────────────────────────┐
│  Security Layers                           │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  1. Transport Security             │   │
│  │  ├─ HTTPS/TLS encryption           │   │
│  │  ├─ Certificate validation         │   │
│  │  └─ Certificate pinning (optional) │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  2. Authentication                 │   │
│  │  ├─ OIDC token validation          │   │
│  │  ├─ Token expiration check         │   │
│  │  └─ GitHub webhook signature       │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  3. Authorization                  │   │
│  │  ├─ User/app permissions           │   │
│  │  ├─ Rate limiting per user         │   │
│  │  └─ Agent access control (future)  │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  4. Input Validation               │   │
│  │  ├─ Request body validation        │   │
│  │  ├─ Agent codename verification    │   │
│  │  ├─ Message format validation      │   │
│  │  └─ SQL injection prevention       │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  5. Output Encoding                │   │
│  │  ├─ JSON encoding                  │   │
│  │  ├─ HTML entity escaping           │   │
│  │  └─ Response sanitization          │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  6. Audit & Monitoring             │   │
│  │  ├─ Request logging                │   │
│  │  ├─ Security event logging         │   │
│  │  ├─ Performance monitoring         │   │
│  │  └─ Anomaly detection              │   │
│  └────────────────────────────────────┘   │
│                                            │
└────────────────────────────────────────────┘
```

## Performance Optimization

### Caching Strategy

```
┌─────────────────────────────────────────┐
│  Multi-Level Caching                    │
│                                         │
│  Level 1: In-Memory Registry            │
│  ├─ Agent metadata (always cached)      │
│  ├─ Handler instances (always cached)   │
│  └─ TTL: None (statically loaded)       │
│                                         │
│  Level 2: Request Cache (Redis)         │
│  ├─ Successful responses                │
│  ├─ TTL: 1-24 hours (configurable)      │
│  └─ Key: hash(agent+messages+context)   │
│                                         │
│  Level 3: Experience Cache (Memory)     │
│  ├─ Retrieved experiences               │
│  ├─ TTL: Depends on fitness score       │
│  └─ Eviction: LRU when limit reached    │
│                                         │
└─────────────────────────────────────────┘
```

### Optimization Techniques

1. **Agent Loading Optimization**
   - Parallel file I/O where possible
   - Cache YAML parsing results
   - Pre-compile regex patterns
   - Lazy load agent details

2. **Memory System Optimization**
   - Sub-linear retrieval (O(1) to O(log n))
   - Bloom filters for quick rejection
   - LSH for approximate matching
   - HNSW for semantic search

3. **Request Processing Optimization**
   - Connection pooling
   - Request queuing
   - Goroutine reuse
   - Buffer pooling

## Monitoring & Observability

```
┌────────────────────────────────────────┐
│  Observability Stack                   │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  Metrics (Prometheus)            │  │
│  │  ├─ Request rate                 │  │
│  │  ├─ Response time (p50/p99)      │  │
│  │  ├─ Error rate by agent          │  │
│  │  └─ Memory/CPU usage             │  │
│  └──────────────────────────────────┘  │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  Logging (Structured)            │  │
│  │  ├─ Request logging              │  │
│  │  ├─ Error logging                │  │
│  │  ├─ Performance logging          │  │
│  │  └─ Security event logging       │  │
│  └──────────────────────────────────┘  │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  Tracing (OpenTelemetry)         │  │
│  │  ├─ Request tracing              │  │
│  │  ├─ Agent invocation tracing     │  │
│  │  ├─ Memory operation tracing     │  │
│  │  └─ Dependency tracing           │  │
│  └──────────────────────────────────┘  │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  Dashboards (Grafana)            │  │
│  │  ├─ System health                │  │
│  │  ├─ Agent performance            │  │
│  │  ├─ Memory system health         │  │
│  │  └─ Error patterns               │  │
│  └──────────────────────────────────┘  │
│                                        │
└────────────────────────────────────────┘
```

---

**For more details, see:**
- [Developer Guide](DEVELOPER_GUIDE.md) - Implementation details
- [Agent Loading Guide](AGENT_LOADING_GUIDE.md) - Agent system
- [API Reference](API_REFERENCE.md) - API endpoints
