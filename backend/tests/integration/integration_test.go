//go:build integration
// +build integration

// Package integration provides end-to-end integration tests for the Elite Agent Collective.
package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

// testServer is the shared test server for all integration tests.
var testServer *httptest.Server

// testRegistry is the shared agent registry for tests.
var testRegistry *agents.Registry

// TestMain sets up and tears down the test server for all integration tests.
func TestMain(m *testing.M) {
	// Initialize agent registry
	testRegistry = agents.DefaultRegistry()

	// Initialize handlers
	agentHandler := agents.NewHandler(testRegistry)

	// Initialize authentication middleware (disabled for most tests)
	cfg := &config.OIDCConfig{
		Issuer:   "https://token.actions.githubusercontent.com",
		ClientID: "", // Empty = auth disabled
	}
	authMiddleware := auth.NewMiddleware(cfg)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Health check endpoint
	r.Get("/health", healthCheckHandler)

	// API routes
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{codename}", agentHandler.GetAgent)
		r.With(authMiddleware.Authenticate).Post("/{codename}/invoke", agentHandler.InvokeAgent)
	})

	// Copilot webhook endpoint
	r.With(authMiddleware.Authenticate).Post("/copilot", agentHandler.CopilotWebhook)

	// Create test server
	testServer = httptest.NewServer(r)

	// Run tests
	code := m.Run()

	// Close the test server
	testServer.Close()

	os.Exit(code)
}

// healthCheckHandler is a simple health check for the test server.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","service":"elite-agent-collective-test"}`))
}

// getTestServerURL returns the URL of the test server.
func getTestServerURL() string {
	return testServer.URL
}
