# 🧠 MNEMONIC: Multi-Agent Neural Experience Memory with Optimized Sub-Linear Inference for Collectives

> A custom Evo-Memory and ReMem-inspired framework designed specifically for the Elite Agent Collective

## Overview

MNEMONIC is a specialized experience memory system that enables the 40 Elite Agents to:

1. **Accumulate and evolve strategies** across task streams without retraining
2. **Share cross-agent experiences** through a collective memory pool
3. **Achieve sub-linear retrieval** using advanced indexing techniques (O(log n) or O(1) lookups)
4. **Self-improve at inference time** using the ReMem-style Think-Act-Refine loop

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    MNEMONIC: Elite Collective Memory System                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                    TIER 0: GLOBAL MEMORY ORCHESTRATOR                    │    │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌─────────────┐    │    │
│  │  │  LSH Index   │ │ Bloom Filter │ │ Hierarchical │ │ Experience  │    │    │
│  │  │ (O(1) lookup)│ │ (membership) │ │  HNSW Graph  │ │ Compressor  │    │    │
│  │  └──────────────┘ └──────────────┘ └──────────────┘ └─────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                    │                                             │
│                ┌───────────────────┼───────────────────┐                        │
│                ▼                   ▼                   ▼                        │
│  ┌────────────────────┐ ┌────────────────────┐ ┌────────────────────┐          │
│  │  AGENT-LOCAL MEM   │ │  TIER-SHARED MEM   │ │  COLLECTIVE MEM    │          │
│  │  ────────────────  │ │  ────────────────  │ │  ────────────────  │          │
│  │  @APEX experiences │ │  Tier-1 strategies │ │  Cross-tier        │          │
│  │  @CIPHER patterns  │ │  Tier-2 patterns   │ │  breakthrough      │          │
│  │  @VELOCITY optims  │ │  Domain knowledge  │ │  discoveries       │          │
│  └────────────────────┘ └────────────────────┘ └────────────────────┘          │
│                                                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                         ReMem-ELITE CONTROL LOOP                                 │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐       │
│  │ RETRIEVE│───▶│  THINK  │───▶│   ACT   │───▶│ REFLECT │───▶│ EVOLVE  │       │
│  │ O(log n)│    │ Reason  │    │ Execute │    │Evaluate │    │ Update  │       │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Sub-Linear Memory Retrieval System

Located in `backend/internal/memory/sublinear_retriever.go`

The retrieval system combines three sub-linear data structures:

#### Bloom Filter (O(1) membership testing)
- Used for quick exact-match checks before expensive operations
- Configured for 1M items with 1% false positive rate
- Enables instant determination if a task signature exists

#### LSH Index (O(1) expected approximate nearest neighbor)
- Uses random hyperplane hashing for cosine similarity
- 10 hash tables with 12 hash functions per table
- Provides constant-time approximate matching

#### HNSW Graph (O(log n) hierarchical search)
- Hierarchical Navigable Small World graph
- M=16 connections, efConstruction=200
- Fallback for semantic search when LSH doesn't find matches

### 2. ReMem-Elite Control Loop

Located in `backend/internal/memory/remem_loop.go`

The control loop implements five phases:

```
RETRIEVE → THINK → ACT → REFLECT → EVOLVE
```

1. **RETRIEVE**: Sub-linear experience retrieval using the tiered approach
2. **THINK**: Build augmented context with retrieved experiences
3. **ACT**: Execute agent with memory-augmented context
4. **REFLECT**: Evaluate outcome using heuristic-based assessment
5. **EVOLVE**: Update memory and fitness scores based on outcome

### 3. Experience Storage

Located in `backend/internal/memory/experience.go`

Each experience tuple captures:
- Task signature (for exact matching)
- Input/output pairs
- Strategy description
- Success status and fitness score
- Embedding vector for semantic search
- Evolution generation tracking
- Usage statistics

### 4. Collective Memory

The system supports three levels of memory sharing:

1. **Agent-Local**: Experiences specific to each agent
2. **Tier-Shared**: Experiences shared among same-tier agents
3. **Collective Breakthroughs**: High-performing experiences shared across all tiers

## Configuration

Configuration is defined in `backend/pkg/models/memory.go`:

```go
MemoryConfig{
    Enabled:                  true,
    EmbeddingDimension:       384,
    MaxExperiencesPerAgent:   1000,
    MinFitnessThreshold:      0.3,
    BreakthroughThreshold:    0.9,
    EvolutionIntervalSeconds: 3600,
    LSHNumTables:             10,
    LSHNumHashFuncs:          12,
    HNSWMaxConnections:       16,
    HNSWEfConstruction:       200,
    HNSWEfSearch:             100,
}
```

## Agent-Specific Memory Roles

| Agent | Memory Specialization |
|-------|----------------------|
| **@OMNISCIENT** | Global memory orchestrator, evolution coordinator |
| **@VELOCITY** | Sub-linear retrieval optimization, algorithm selection |
| **@TENSOR** | Embedding generation, similarity computation |
| **@NEURAL** | Experience synthesis, pattern recognition |
| **@PRISM** | Fitness statistics, success rate analysis |
| **@NEXUS** | Cross-domain experience transfer |
| **@GENESIS** | Novel strategy mutation and exploration |

## Performance Targets

- **Retrieval latency**: < 5ms for 1M experiences (sub-linear guarantee)
- **Memory footprint**: Compressed experiences reduce storage by ~60%
- **Evolution cycle**: Background evolution every 1000 interactions
- **Cross-agent transfer**: Real-time breakthrough propagation

## Key Innovations Over Standard Evo-Memory/ReMem

| Feature | Standard Evo-Memory | MNEMONIC (Elite) |
|---------|---------------------|------------------|
| **Retrieval Complexity** | O(n) or O(n log n) | **O(1) - O(log n)** via LSH/HNSW |
| **Agent Scope** | Single agent | **40 collaborative agents** |
| **Memory Sharing** | Isolated | **Tier-based + collective pool** |
| **Evolution Trigger** | Per-task | **Multi-level: agent, tier, collective** |
| **Strategy Synthesis** | Basic | **Cross-domain synthesis via @NEXUS** |
| **Orchestration** | None | **@OMNISCIENT meta-coordination** |

## Directory Structure

```
backend/
├── internal/
│   └── memory/
│       ├── experience.go          # ExperienceTuple definitions
│       ├── sublinear_retriever.go # LSH, HNSW, Bloom filters
│       ├── remem_loop.go          # ReMem control loop
│       └── errors.go              # Error definitions
└── pkg/
    └── models/
        └── memory.go              # Shared memory data models
```

## Usage Example

```go
import (
    "github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
    "github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// Initialize the ReMem controller
config := memory.DefaultReMemConfig()
embeddingService := memory.NewNoOpEmbeddingService(config.EmbeddingDimension)
controller := memory.NewReMemController(config, embeddingService)

// Execute with memory augmentation
response, err := controller.ExecuteWithMemory(
    ctx,
    "APEX",          // Agent ID
    request,         // *models.CopilotRequest
    agentExecutor,   // AgentExecutor interface
)
```

## Future Enhancements

1. **Real Embedding Service**: Replace NoOpEmbeddingService with actual embeddings (OpenAI, local models)
2. **Persistence Layer**: Add memory persistence for recovery across restarts
3. **Evolution Engine**: Implement the full evolution engine for strategy synthesis
4. **Distributed Memory**: Scale across multiple instances with shared memory pool
5. **Memory Compression**: Implement information-theoretic compression for long-term storage

## Related Documentation

- [Sub-Linear Techniques](./sub-linear-techniques.md)
- [Experience Evolution](./experience-evolution.md)
- [API Reference](../api-reference/endpoints.md)
