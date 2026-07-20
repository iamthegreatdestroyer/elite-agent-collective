// Package main is the entry point for the Elite Agent Collective backend server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/retrieval"
)

// corsMiddleware creates CORS middleware with configurable allowed origins.
// If allowedOrigins is empty, it allows all origins (for development).
// In production, set CORS_ALLOWED_ORIGINS to restrict to specific domains.
func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := allowedOrigins
			if origin == "" {
				// Default to allowing all origins if not configured
				// For production, set CORS_ALLOWED_ORIGINS environment variable
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-GitHub-Signature-256")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize upstream proxy client
	upstreamClient := copilot.NewUpstreamClient()
	if upstreamClient.Enabled() {
		log.Printf("Upstream proxy enabled: forwarding to %s", os.Getenv("COPILOT_UPSTREAM_URL"))
	} else {
		log.Printf("Upstream proxy disabled (set COPILOT_UPSTREAM_URL to enable)")
	}

	// Initialize local Ollama fallback client (second-tier, used when upstream
	// Copilot is disabled or fails). Configured via OLLAMA_URL (default
	// http://localhost:8000) and OLLAMA_MODEL (default phi4-mini).
	ollamaClient := copilot.NewOllamaClient()
	if ollamaClient.Enabled() {
		log.Printf("Ollama fallback enabled: %s (model %s)", ollamaClient.BaseURL(), ollamaClient.Model())
		// Pre-load the model in the background so the first real Copilot
		// request does not pay the CPU cold-load penalty (~40-50s) and time out.
		go ollamaClient.Warmup(context.Background())
	} else {
		log.Printf("Ollama fallback unavailable at %s (agents will use template responses if upstream is also unavailable)", ollamaClient.BaseURL())
	}

	// Initialize persistent memory store (Mem0 pattern).
	dataDir := envStr("MEMORY_DATA_DIR", filepath.Join("data", "memories"))
	memStore, err := memory.NewStore(dataDir)
	if err != nil {
		log.Printf("Warning: could not initialize memory store (%v) — running without persistence", err)
		memStore = nil
	} else {
		log.Printf("Memory store initialized at %s", dataDir)
	}

	// Initialize agentmem MCP client (real semantic-memory backend for
	// Remember/Recall - github.com/iamthegreatdestroyer/agentmem - replacing
	// the local per-user JSON files when configured). Configured via
	// AGENTMEM_MCP_URL; unset means Remember/Recall keep using the local
	// JSON files exactly as before.
	agentMemClient := memory.NewAgentMemClient(envStr("AGENTMEM_MCP_URL", ""))
	if agentMemClient.Enabled() {
		if agentMemClient.Healthy() {
			log.Printf("agentmem semantic memory enabled: %s", agentMemClient.BaseURL())
		} else {
			log.Printf("Warning: AGENTMEM_MCP_URL=%s is set but not responding — Remember/Recall will error until it is reachable", agentMemClient.BaseURL())
		}
	} else {
		log.Printf("agentmem not configured (set AGENTMEM_MCP_URL to enable); using local JSON fact store")
	}
	if memStore != nil {
		memStore.SetAgentMem(agentMemClient)
	}

	// Embedding client for semantic memory recall (gateway /api/embed,
	// nomic-embed-text). Default-on but fully fail-open: if the gateway is
	// down, RecallRelevant silently degrades to recency ordering.
	embedClient := memory.NewEmbedClient(envStr("EMBED_URL", "http://localhost:8000"), envStr("EMBED_MODEL", "nomic-embed-text"))
	if embedClient.Enabled() {
		log.Printf("Semantic memory recall enabled via %s", embedClient.BaseURL())
	} else {
		log.Printf("Semantic memory recall disabled (set EMBED_URL); using recency ordering")
	}
	if memStore != nil {
		memStore.SetEmbedder(embedClient)
	}

	// Initialize agent registry
	registry := agents.DefaultRegistryWithUpstream(upstreamClient)
	registry.InjectOllama(ollamaClient)
	if retriever := retrieval.NewSigmaIndexRetriever(); retriever != nil {
		registry.InjectRetriever(retriever)
		if retriever.Enabled() {
			log.Printf("Retriever enabled: sigma-index %s", retriever.IndexURL())
		} else {
			log.Printf("Retriever configured but sigma-index %s not reachable yet (fail-open per request)", retriever.IndexURL())
		}
	} else {
		log.Printf("Retriever disabled (set EAC_RETRIEVER=1 to augment prompts from in-my-head via sigma-index)")
	}
	if memStore != nil {
		registry.InjectMemory(memStore)
	}
	log.Printf("Registered %d agents", registry.Count())

	// Initialize crew registry (CrewAI pattern).
	crewsPath := envStr("CREWS_CONFIG", filepath.Join("..", "..", "config", "crews.yaml"))
	crewRegistry, err := agents.NewCrewRegistry(crewsPath)
	if err != nil {
		log.Printf("Warning: could not load crews from %s (%v)", crewsPath, err)
		crewRegistry, _ = agents.NewCrewRegistry("") // empty registry
	} else {
		log.Printf("Loaded %d crew definitions from %s", crewRegistry.Count(), crewsPath)
	}

	// Initialize handlers
	agentHandler := agents.NewHandler(registry).WithCrews(crewRegistry)
	registry.InjectCrews(crewRegistry)

	// Initialize authentication middleware
	authMiddleware := auth.NewMiddleware(&cfg.OIDC)

	// Initialize signature verification middleware for GitHub webhooks
	signatureMiddleware := auth.NewSignatureMiddleware(cfg.GitHub.WebhookSecret)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(metrics.Middleware)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware(cfg.CORSAllowedOrigins))

	// Health check endpoint (no auth required)
	r.Get("/health", healthCheckHandler)
	r.Get("/metrics", metrics.Handler)

	// API routes
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{codename}", agentHandler.GetAgent)
		r.With(authMiddleware.Authenticate).Post("/{codename}/invoke", agentHandler.InvokeAgent)
	})

	// Task queue proxy: forward /tasks to the eac-taskd worker (runs as an
	// unprivileged sidecar that owns the SQLite queue, reaches the Ryzanstein
	// gateway, and writes the in-my-head vault). This backend stays root +
	// sandboxed and never does LLM work itself; it only enqueues via HTTP.
	// POST /tasks returns 202 immediately (enqueue), so the 60s router timeout
	// is never near the multi-minute dispatch. Fail-open: 502 + JSON if the
	// worker is unreachable, leaving every other route unaffected.
	if taskdURL := envStr("EAC_TASKD_URL", "http://127.0.0.1:29081"); taskdURL != "" {
		if target, err := url.Parse(taskdURL); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, e error) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				fmt.Fprintf(w, `{"error":"eac-taskd worker unreachable at %s","detail":%q}`+"\n", taskdURL, e.Error())
			}
			r.Handle("/tasks", proxy)
			r.Handle("/tasks/*", proxy)
			log.Printf("Task queue proxy enabled: /tasks -> %s (eac-taskd worker)", taskdURL)
		} else {
			log.Printf("Task queue proxy disabled: bad EAC_TASKD_URL %q (%v)", taskdURL, err)
		}
	}

	// Copilot webhook endpoint with signature verification
	// Uses signature verification when GITHUB_WEBHOOK_SECRET is configured
	// Falls back to OIDC auth otherwise
	r.With(signatureMiddleware.VerifySignature, authMiddleware.OptionalAuth).Post("/copilot", agentHandler.CopilotWebhook)

	// Alternative Copilot endpoint with only OIDC auth (for direct API calls)
	r.With(authMiddleware.Authenticate).Post("/agent", agentHandler.CopilotWebhook)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	log.Printf("Server is starting on %s", addr)
	log.Printf("Health check available at http://localhost%s/health", addr)
	log.Printf("Agent list available at http://localhost%s/agents", addr)
	log.Printf("Copilot webhook at http://localhost%s/copilot", addr)

	if cfg.GitHub.WebhookSecret != "" {
		log.Printf("GitHub webhook signature verification enabled")
	}
	if cfg.OIDC.ClientID != "" {
		log.Printf("OIDC authentication enabled")
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", addr, err)
	}

	<-done
	log.Println("Server stopped")
}

// healthCheckHandler handles the /health endpoint.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "elite-agent-collective",
		"version":   "3.0.0",
	}
	json.NewEncoder(w).Encode(response)
}

// envStr reads a string env var, returning def if not set.
func envStr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
