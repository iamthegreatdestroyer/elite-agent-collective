package retrieval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubBackends spins up a gateway (/api/embed) and a sigma-index (/health,
// /search) that returns the REAL capitalized wire shape including Text.
func stubBackends(t *testing.T) (gateway, index *httptest.Server) {
	t.Helper()
	gateway = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"model":"nomic-embed-text","embeddings":[[0.1,0.2,0.3]]}`))
	}))
	index = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(200)
		case "/search":
			// Capitalized untagged keys, exactly like the Go server encodes.
			_, _ = w.Write([]byte(`{"count":2,"results":[` +
				`{"ID":"note-vault","Score":7.19,"VecDist":0,"TextRank":1,"Text":"vault-git provides AES-256-GCM encrypted snapshots of the second brain."},` +
				`{"ID":"note-kyber","Score":3.08,"VecDist":0,"TextRank":2,"Text":"sigmavault uses Kyber-1024 and Dilithium-3 post-quantum crypto."}` +
				`]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return gateway, index
}

func TestRetriever_Disabled(t *testing.T) {
	// EAC_RETRIEVER unset -> nil (opt-in).
	t.Setenv("EAC_RETRIEVER", "")
	if r := NewSigmaIndexRetriever(); r != nil {
		t.Fatal("expected nil retriever when EAC_RETRIEVER unset")
	}
	t.Setenv("EAC_RETRIEVER", "off")
	if r := NewSigmaIndexRetriever(); r != nil {
		t.Fatal("expected nil retriever when EAC_RETRIEVER=off")
	}
}

func TestRetriever_Retrieve(t *testing.T) {
	gw, idx := stubBackends(t)
	defer gw.Close()
	defer idx.Close()

	t.Setenv("EAC_RETRIEVER", "1")
	t.Setenv("EAC_GATEWAY_URL", gw.URL)
	t.Setenv("EAC_SIGMA_INDEX_URL", idx.URL)

	r := NewSigmaIndexRetriever()
	if r == nil {
		t.Fatal("retriever should be enabled")
	}
	if !r.Enabled() {
		t.Fatal("Enabled() should be true (health 200)")
	}
	block, err := r.Retrieve(context.Background(), "Advanced Cryptography & Security", "how does the vault encrypt snapshots")
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	for _, want := range []string{"in-my-head knowledge base", "note-vault", "AES-256-GCM", "note-kyber", "Kyber-1024"} {
		if !strings.Contains(block, want) {
			t.Fatalf("block missing %q:\n%s", want, block)
		}
	}
}

func TestRetriever_FailOpen(t *testing.T) {
	t.Setenv("EAC_RETRIEVER", "1")
	t.Setenv("EAC_GATEWAY_URL", "http://127.0.0.1:1")   // unreachable
	t.Setenv("EAC_SIGMA_INDEX_URL", "http://127.0.0.1:1")
	r := NewSigmaIndexRetriever()
	if r == nil {
		t.Fatal("retriever should be constructed")
	}
	if r.Enabled() {
		t.Fatal("Enabled() should be false when /health is unreachable")
	}
	if _, err := r.Retrieve(context.Background(), "x", "y"); err == nil {
		t.Fatal("Retrieve should error (fail-open) when backends are down")
	}
}
