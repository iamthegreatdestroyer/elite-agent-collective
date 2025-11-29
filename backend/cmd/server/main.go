// Package main is the entry point for the Elite Agent Collective backend server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize agent registry
	registry := agents.DefaultRegistry()
	log.Printf("Registered %d agents", registry.Count())

	// Initialize handlers
	agentHandler := agents.NewHandler(registry)

	// Initialize authentication middleware
	authMiddleware := auth.NewMiddleware(&cfg.OIDC)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint (no auth required)
	r.Get("/health", healthCheckHandler)

	// API routes
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{codename}", agentHandler.GetAgent)
		r.With(authMiddleware.Authenticate).Post("/{codename}/invoke", agentHandler.InvokeAgent)
	})

	// Copilot webhook endpoint
	r.With(authMiddleware.Authenticate).Post("/copilot", agentHandler.CopilotWebhook)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
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
		"version":   "1.0.0",
	}
	json.NewEncoder(w).Encode(response)
}
