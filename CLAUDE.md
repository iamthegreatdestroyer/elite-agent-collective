# elite-agent-collective — Make Agents Real

## Current State: TEMPLATE RESPONSES
Agents return hardcoded templates when COPILOT_UPSTREAM_URL is not set.
Memory is keyword-based, not vector/semantic. No local LLM fallback.

## Sprint 1: Ollama Fallback
- [ ] In backend/internal/upstream/upstream.go, add Ollama fallback
- [ ] When COPILOT_UPSTREAM_URL is empty or fails, call http://localhost:11434/api/chat
- [ ] Use model from OLLAMA_MODEL env var, default phi4-mini
- [ ] Format: prepend agent system prompt, send user message, stream response
- [ ] Test: start without COPILOT_UPSTREAM_URL, verify agents respond with real content

## Sprint 2: Vector Memory
- [ ] Replace keyword extraction in memory.Store with embedding-based storage
- [ ] Call Ryzanstein /v1/embeddings for embedding generation
- [ ] Store embeddings in-memory with cosine similarity search
- [ ] FormatContext returns top-5 most relevant memories

## Sprint 3: Parallel Pipeline
- [ ] In pipeline.go, run non-dependent agents concurrently using goroutines
- [ ] Only serialize when one agent output feeds another
- [ ] Add timeout per agent, 30 seconds default

## Build Commands
```bash
export PATH=$PATH:/usr/local/go/bin
cd /opt/sigmavault/repos/meta-elite-agent-collective
go build ./...
go test ./...
```

## Done Criteria
- [ ] Agents produce real LLM output without upstream URL
- [ ] Memory uses vector similarity
- [ ] go build succeeds
- [ ] go test passes

## Completion Signal
```bash
git tag v4.0.0
```
