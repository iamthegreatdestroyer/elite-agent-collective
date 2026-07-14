# elite-agent-collective — Make Agents Real

## Current State: REAL OLLAMA FALLBACK WIRED (2026-07-03)
Agents forward to upstream Copilot when COPILOT_UPSTREAM_URL is set; when it
is unset or the upstream call fails, they now fall through to a local Ollama
model (backend/internal/copilot/ollama_client.go) as a second tier. Only if
BOTH upstream and Ollama are unavailable do agents fall back to the hardcoded
template response. Memory fact STORAGE and RECALL are embedding-backed:
Remember() embeds each fact via the Ryzanstein gateway (/api/embed,
nomic-embed-text) and caches the vector on the Fact; RecallRelevant() ranks
facts by cosine similarity (threshold 0.55) with a recency fallback, wired
through SetEmbedder(NewEmbedClient) in main.go and consumed via
FormatContextRelevant in both handlers (apex.go, base.go). Keyword matching
(memory.ExtractFacts) now only gates WHICH text becomes a fact, not how facts
are stored or recalled. The semantic_network.go spreading-activation network
remains scaffolding, never wired — deferred to v4.1+ (see Future Work below).

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
- NOTE: the "embedding-based storage" above refers to the SHIPPED
      embedding-backed fact store + recall (nomic-embed via /api/embed, cosine
      threshold 0.55 with recency fallback), which is live today and is the
      decided v4.0.0 memory bar (see Done Criteria). The spreading-activation
      semantic_network.go is NOT wired (only its own tests reference it) and is
      deferred to v4.1+ (see Future Work). The Sprint-2 [x] and the Done
      Criterion now agree: embedding-backed recall is the v4.0.0 memory bar.

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
- [x] Memory recall is embedding-backed — facts embedded at Remember time
      (nomic-embed-text via gateway /api/embed), retrieved by cosine
      similarity (RecallRelevant, threshold 0.55) with recency fallback, wired
      via SetEmbedder(NewEmbedClient) in main.go and consumed by both handlers
      (apex.go, base.go) via FormatContextRelevant. This is the decided v4.0.0
      memory bar (user decision 2026-07-14). The XL spreading-activation
      semantic network (semantic_network.go) is explicitly OUT of scope for
      v4.0.0 and deferred to v4.1+ (see Future Work).
- [x] go build succeeds
- [x] go test ./... passes, 0 failures (was 2 real failures: a flaky
      population-generation bug in architecture_search.go and a premature
      decay-eviction bug in cognitive_working_memory.go — both root-caused
      and fixed 2026-07-03, not test off-by-ones)

## Future Work (v4.1+)
- [ ] Spreading-activation semantic network (semantic_network.go): the XL
      cognitive-architecture recall model (associative spreading activation
      beyond flat vector cosine similarity). Currently scaffolding —
      referenced only by its own tests, never wired into Remember/Recall or
      the handlers. Explicitly OUT of scope for v4.0.0; a candidate for a
      future major memory upgrade.

## Completion Signal
```bash
git tag v4.0.0
```
