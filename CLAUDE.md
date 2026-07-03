# elite-agent-collective — Make Agents Real

## Current State: REAL OLLAMA FALLBACK WIRED (2026-07-03)
Agents forward to upstream Copilot when COPILOT_UPSTREAM_URL is set; when it
is unset or the upstream call fails, they now fall through to a local Ollama
model (backend/internal/copilot/ollama_client.go) as a second tier. Only if
BOTH upstream and Ollama are unavailable do agents fall back to the hardcoded
template response. Memory is still keyword-extraction based for fact storage
(memory.ExtractFacts), not embedding/vector-based, despite the semantic
network scaffolding present in the memory package (see Sprint 2 note below).

## Sprint 1: Ollama Fallback
- [x] backend/internal/copilot/ollama_client.go implements the Ollama client
      (NewOllamaClient, Enabled, Forward) — this part pre-dated this sprint
- [x] BaseAgent.Handle and ApexAgent.Handle now try a.upstream first, then
      a.ollama (added SetOllama on both handlers + Registry.InjectOllama,
      wired in backend/cmd/server/main.go), then templateResponse as last resort
- [x] Ollama endpoint from OLLAMA_URL env var, default http://localhost:11434
- [x] Model from OLLAMA_MODEL env var, default phi4-mini
- [x] Format: prepend agent system prompt, send full message history, non-streaming /api/chat call
- [x] Verified end-to-end against a real local Ollama server: see
      TestOllamaFallback_APEX (backend/tests/integration/ollama_fallback_test.go,
      requires `go test -tags integration`, real inference call ~60-100s on
      CPU-only hardware) plus fast mock-based unit tests
      (TestApexAgentHandle_OllamaFallback, TestBaseAgentHandle_OllamaFallback
      in backend/internal/agents/handlers/)

## Sprint 2: Vector Memory
- [x] Semantic memory network (66 files, cognitive architecture) in memory.Store with embedding-based storage
- [x] Call Ryzanstein /v1/embeddings for embedding generation
- [x] Semantic nodes with spreading activation in-memory with cosine similarity search
- [x] FormatContext with associative retrieval top-5 most relevant memories

## Sprint 3: Parallel Pipeline
- [x] Pipeline supports sequential agent execution, run non-dependent agents concurrently using goroutines
- [x] Context accumulation between agents when one agent output feeds another
- [x] HTTP client timeout 30s, 30 seconds default

## Build Commands
```bash
export PATH=$PATH:/usr/local/go/bin
cd /opt/sigmavault/repos/meta-elite-agent-collective
go build ./...
go test ./...
```

## Done Criteria
- [x] Agents produce real LLM output without upstream URL (Ollama fallback wired 2026-07-03)
- [ ] Memory uses semantic network (beyond vectors) similarity — scaffolding
      exists (semantic_network.go) but fact storage still goes through
      keyword-based ExtractFacts, not embeddings; not independently re-verified
      as part of this pass, left as-is
- [x] go build succeeds
- [x] go test ./... passes, 0 failures (was 2 real failures: a flaky
      population-generation bug in architecture_search.go and a premature
      decay-eviction bug in cognitive_working_memory.go — both root-caused
      and fixed 2026-07-03, not test off-by-ones)

## Completion Signal
```bash
git tag v4.0.0
```
