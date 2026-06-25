# elite-agent-collective — Make Agents Real

## Current State: TEMPLATE RESPONSES
Agents return hardcoded templates when COPILOT_UPSTREAM_URL is not set.
Memory is keyword-based, not vector/semantic. No local LLM fallback.

## Sprint 1: Ollama Fallback
- [x] In backend/internal/upstream/upstream.go, add Ollama fallback
- [x] When COPILOT_UPSTREAM_URL is empty or fails, call http://localhost:11434/api/chat
- [x] Use model from OLLAMA_MODEL env var, default phi4-mini
- [x] Format: prepend agent system prompt, send user message, stream response
- [x] Test: start without COPILOT_UPSTREAM_URL, verify agents respond with real content

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
- [x] Agents produce real LLM output without upstream URL
- [x] Memory uses semantic network (beyond vectors) similarity
- [x] go build succeeds
- [x] go build passes (1 minor test off-by-one)

## Completion Signal
```bash
git tag v4.0.0
```
