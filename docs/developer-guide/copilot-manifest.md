# Copilot Extension Manifest Guide

This guide explains the structure and maintenance of the Elite Agent Collective's Copilot Extension manifest files.

## Overview

The Elite Agent Collective uses a manifest-based approach to define all 40 agents. This enables:

- **Consistency**: Single source of truth for agent definitions
- **Automation**: Auto-generation of the Copilot extension configuration
- **Validation**: Programmatic verification of manifest completeness
- **Extensibility**: Easy addition of new agents

## File Structure

```
elite-agent-collective/
├── copilot-extension.json     # Generated Copilot extension manifest
├── config/
│   └── agents-manifest.yaml   # Source of truth for agent definitions
├── scripts/
│   ├── generate-copilot-manifest.go
│   └── go.mod
└── backend/
    └── internal/agents/
        └── registry.go        # Can load from manifest
```

## Manifest Files

### agents-manifest.yaml

The YAML manifest is the **source of truth** for all agent definitions. It contains:

```yaml
version: "2.0"
name: "Elite Agent Collective"
description: "A system of 40 specialized AI agents..."

tiers:
  - id: 1
    name: "Foundational"
    description: "Core engineering disciplines"
  # ... all 8 tiers

agents:
  - id: "01"
    codename: "APEX"
    tier: 1
    name: "Elite Computer Science Engineering Specialist"
    description: "Master-level software engineering..."
    philosophy: "Every problem has an elegant solution..."
    keywords:
      - "software engineering"
      - "algorithms"
    directives:
      - "Deliver production-grade code"
      - "Apply CS fundamentals deeply"
    examples:
      - "@APEX implement rate limiter"
      - "@APEX analyze algorithm"
    collaborators:
      - "VELOCITY"
      - "ARCHITECT"
```

#### Agent Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Two-digit numeric ID (01-40) |
| `codename` | Yes | Agent name in UPPERCASE |
| `tier` | Yes | Tier number (1-8) |
| `name` | Yes | Full agent title/specialty |
| `description` | Yes | One-sentence capability description |
| `philosophy` | Yes | Guiding principle quote |
| `keywords` | Yes | Activation keywords for auto-routing |
| `directives` | Yes | Core behavioral directives (3-5) |
| `examples` | Yes | Example invocation commands |
| `collaborators` | Yes | List of complementary agent codenames |

### copilot-extension.json

The JSON manifest is the **Copilot Extension configuration** used by GitHub Copilot. It is generated from `agents-manifest.yaml`.

```json
{
  "$schema": "https://json.schemastore.org/github-copilot-extension.json",
  "name": "elite-agent-collective",
  "display_name": "Elite Agent Collective",
  "version": "2.0.0",
  "tools": [
    {
      "name": "APEX",
      "description": "Elite Computer Science Engineering Specialist...",
      "parameters": {
        "type": "object",
        "properties": {
          "task": {
            "type": "string",
            "description": "The task for APEX to perform..."
          }
        },
        "required": ["task"]
      }
    }
  ],
  "tiers": {
    "1": {
      "name": "Foundational",
      "agents": ["APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY"]
    }
  }
}
```

## Generating the Manifest

### Prerequisites

- Go 1.21 or later

### Steps

1. **Navigate to the scripts directory**:
   ```bash
   cd scripts
   ```

2. **Build the generator** (first time only):
   ```bash
   go build -o generate-manifest ./generate-copilot-manifest.go
   ```

3. **Run the generator**:
   ```bash
   ./generate-manifest
   ```

4. **Verify the output**:
   ```bash
   cat ../copilot-extension.json | jq '.tools | length'
   # Should output: 40
   ```

### Validation

The generator automatically validates:
- ✅ Exactly 40 agents are defined
- ✅ No duplicate codenames
- ✅ All required fields are present
- ✅ Correct tier distribution (5+12+2+1+5+5+5+5 = 40)

## Adding a New Agent

### Step 1: Update agents-manifest.yaml

Add the new agent definition in the appropriate tier section:

```yaml
- id: "41"
  codename: "NEWAGENT"
  tier: 6  # Choose appropriate tier
  name: "New Specialty Specialist"
  description: "Description of what this agent does"
  philosophy: "Guiding principle for this agent."
  keywords:
    - "keyword1"
    - "keyword2"
  directives:
    - "First core directive"
    - "Second core directive"
  examples:
    - "@NEWAGENT example task 1"
    - "@NEWAGENT example task 2"
  collaborators:
    - "APEX"
    - "OTHER_AGENT"
```

### Step 2: Update Tier Counts

If adding a 41st agent, update the validation in:
- `scripts/generate-copilot-manifest.go` - Update expected counts
- `backend/internal/agents/registry.go` - Update validation logic

### Step 3: Add Backend Handler

Create or update the agent handler in `backend/internal/agents/`:

```go
// In handlers/newagent.go
type NewAgentHandler struct {
    BaseAgent
}

func NewNewAgentHandler() *NewAgentHandler {
    return &NewAgentHandler{
        BaseAgent: BaseAgent{
            info: models.Agent{...},
        },
    }
}
```

### Step 4: Regenerate Manifest

```bash
cd scripts
./generate-manifest
```

### Step 5: Verify

```bash
# Check JSON is valid
cat copilot-extension.json | jq .

# Run tests
cd backend && go test ./...
```

## Using the Manifest in Backend

The backend can load agents directly from the manifest:

```go
import "github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents"

// Load from manifest file
registry, err := agents.RegistryFromManifest("config/agents-manifest.yaml")
if err != nil {
    log.Fatal(err)
}

// Or use the default (hardcoded) registry
registry := agents.DefaultRegistry()
```

### Validating a Manifest

```go
manifest, err := agents.LoadManifest("config/agents-manifest.yaml")
if err != nil {
    log.Fatal(err)
}

if err := agents.ValidateManifest(manifest); err != nil {
    log.Fatalf("Invalid manifest: %v", err)
}
```

## Tier Reference

| Tier | Name | Count | Description |
|------|------|-------|-------------|
| 1 | Foundational | 5 | Core engineering disciplines |
| 2 | Specialists | 12 | Deep domain expertise |
| 3 | Innovators | 2 | Cross-domain synthesis |
| 4 | Meta | 1 | System orchestration |
| 5 | Domain Specialists | 5 | Infrastructure/data |
| 6 | Emerging Tech | 5 | Cutting-edge technology |
| 7 | Human-Centric | 5 | Developer experience |
| 8 | Enterprise | 5 | Business/compliance |

## Agent Quick Reference

### Tier 1: Foundational
- **APEX** (01) - Computer Science Engineering
- **CIPHER** (02) - Cryptography & Security
- **ARCHITECT** (03) - Systems Architecture
- **AXIOM** (04) - Mathematics & Formal Proofs
- **VELOCITY** (05) - Performance Optimization

### Tier 2: Specialists
- **QUANTUM** (06) - Quantum Computing
- **TENSOR** (07) - Machine Learning
- **FORTRESS** (08) - Defensive Security
- **NEURAL** (09) - Cognitive Computing
- **CRYPTO** (10) - Blockchain
- **FLUX** (11) - DevOps
- **PRISM** (12) - Data Science
- **SYNAPSE** (13) - API Design
- **CORE** (14) - Low-Level Systems
- **HELIX** (15) - Bioinformatics
- **VANGUARD** (16) - Research Analysis
- **ECLIPSE** (17) - Testing & Verification

### Tier 3: Innovators
- **NEXUS** (18) - Paradigm Synthesis
- **GENESIS** (19) - Zero-to-One Innovation

### Tier 4: Meta
- **OMNISCIENT** (20) - Meta-Learning Orchestrator

### Tier 5: Domain Specialists
- **ATLAS** (21) - Cloud Infrastructure
- **FORGE** (22) - Build Systems
- **SENTRY** (23) - Observability
- **VERTEX** (24) - Graph Databases
- **STREAM** (25) - Real-Time Processing

### Tier 6: Emerging Tech
- **PHOTON** (26) - Edge Computing
- **LATTICE** (27) - Distributed Consensus
- **MORPH** (28) - Code Migration
- **PHANTOM** (29) - Reverse Engineering
- **ORBIT** (30) - Embedded Systems

### Tier 7: Human-Centric
- **CANVAS** (31) - UI/UX Design
- **LINGUA** (32) - NLP & LLMs
- **SCRIBE** (33) - Documentation
- **MENTOR** (34) - Code Review
- **BRIDGE** (35) - Cross-Platform

### Tier 8: Enterprise
- **AEGIS** (36) - Compliance
- **LEDGER** (37) - Financial Systems
- **PULSE** (38) - Healthcare IT
- **ARBITER** (39) - Conflict Resolution
- **ORACLE** (40) - Predictive Analytics

## Troubleshooting

### Generator fails with "expected 40 agents"

Check that `agents-manifest.yaml` contains exactly 40 agent definitions. Count with:
```bash
grep "codename:" config/agents-manifest.yaml | wc -l
```

### Generator fails with "duplicate codename"

Each agent must have a unique codename. Search for duplicates:
```bash
grep "codename:" config/agents-manifest.yaml | sort | uniq -d
```

### JSON validation errors

Ensure the generated JSON is valid:
```bash
cat copilot-extension.json | jq . > /dev/null
```

### Tests fail after manifest changes

Run the full test suite:
```bash
cd backend && go test ./...
```

## Related Documentation

- [Adding Agents](adding-agents.md) - Detailed guide for creating new agents
- [Architecture](architecture.md) - System architecture overview
- [Contributing](contributing.md) - Contribution guidelines
