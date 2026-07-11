package memory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbedClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"nomic-embed-text","embeddings":[[0.1,0.2,0.3,0.4]]}`))
	}))
	defer srv.Close()

	c := NewEmbedClient(srv.URL, "")
	if !c.Enabled() {
		t.Fatal("client with a baseURL should be enabled")
	}
	vec, err := c.Embed("hello")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if len(vec) != 4 {
		t.Fatalf("want 4-dim vector, got %d", len(vec))
	}

	// Disabled client: nil-safe Enabled + erroring Embed.
	if NewEmbedClient("", "").Enabled() {
		t.Fatal("empty baseURL should be disabled")
	}
	if _, err := NewEmbedClient("", "").Embed("x"); err == nil {
		t.Fatal("disabled Embed should return an error")
	}
	var nilClient *EmbedClient
	if nilClient.Enabled() {
		t.Fatal("nil client should not be enabled")
	}

	// Non-200 -> error (fail-open contract).
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer bad.Close()
	if _, err := NewEmbedClient(bad.URL, "").Embed("x"); err == nil {
		t.Fatal("non-200 should error")
	}
}

// stubEmbedServer returns a deterministic 3-d vector keyed on the input text so
// tests can control cosine ordering without a live gateway.
func stubEmbedServer(t *testing.T) *httptest.Server {
	t.Helper()
	vecFor := func(s string) []float32 {
		switch {
		case strings.Contains(s, "rust"):
			return []float32{1, 0, 0}
		case strings.Contains(s, "crypto"):
			return []float32{0.95, 0.05, 0}
		case strings.Contains(s, "cooking"):
			return []float32{0, 0, 1}
		default:
			return []float32{0.2, 0.2, 0.2}
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Input string `json:"input"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		v := vecFor(strings.ToLower(body.Input))
		resp, _ := json.Marshal(map[string]interface{}{"embeddings": [][]float32{v}})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}))
}

func TestRecallRelevant_SemanticRanking(t *testing.T) {
	srv := stubEmbedServer(t)
	defer srv.Close()

	st, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	st.SetEmbedder(NewEmbedClient(srv.URL, ""))

	// Facts embedded once at Remember time.
	_ = st.Remember("u1", "APEX", "language", "rust")
	_ = st.Remember("u1", "APEX", "hobby", "cooking")

	// A crypto/rust query should rank the rust fact first, above the newer
	// cooking fact — i.e. relevance beats recency.
	facts := st.RecallRelevant("u1", "building a crypto vault in rust", 1)
	if len(facts) != 1 || facts[0].Key != "language" {
		t.Fatalf("expected 'language' fact ranked first, got %+v", facts)
	}

	// Confirm a vector was persisted at Remember time.
	all := st.Recall("u1")
	for _, f := range all {
		if len(f.Vector) == 0 {
			t.Fatalf("fact %q should have a cached vector", f.Key)
		}
	}
}

func TestRecallRelevant_FailOpenRecency(t *testing.T) {
	st, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// No embedder configured -> pure recency, but never fewer than before.
	_ = st.Remember("u1", "A", "k1", "v1")
	_ = st.Remember("u1", "A", "k2", "v2")

	facts := st.RecallRelevant("u1", "anything", 5)
	if len(facts) != 2 {
		t.Fatalf("expected 2 facts via recency fallback, got %d", len(facts))
	}
	if facts[0].Key != "k2" {
		t.Fatalf("expected newest-first (k2), got %s", facts[0].Key)
	}
}
