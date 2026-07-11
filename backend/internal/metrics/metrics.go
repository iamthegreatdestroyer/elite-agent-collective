// Package metrics provides a tiny, dependency-free Prometheus text-exposition
// for the Elite Agent Collective backend. Scalar counters/gauges are plain
// atomics; labeled counters are mutex-guarded maps. GET /metrics renders the
// standard text/plain; version=0.0.4 format that Prometheus scrapes — no
// prometheus/client_golang dependency is introduced.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Scalar counters and gauges.
var (
	RequestsTotal       atomic.Int64 // counter: all HTTP requests handled
	ErrorsTotal         atomic.Int64 // counter: responses with a 5xx status
	UpstreamTotal       atomic.Int64 // counter: answers served by upstream Copilot
	OllamaFallbackTotal atomic.Int64 // counter: answers served by the local Ollama fallback
	InFlight            atomic.Int64 // gauge: in-flight HTTP requests
)

// labeledCounter is a mutex-guarded label -> count map.
type labeledCounter struct {
	mu   sync.Mutex
	vals map[string]int64
}

func newLabeledCounter() *labeledCounter { return &labeledCounter{vals: map[string]int64{}} }

func (c *labeledCounter) Inc(label string) {
	c.mu.Lock()
	c.vals[label]++
	c.mu.Unlock()
}

func (c *labeledCounter) snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.vals))
	for k, v := range c.vals {
		out[k] = v
	}
	return out
}

var (
	requestsByAgent = newLabeledCounter() // label: agent
	routingMode     = newLabeledCounter() // label: mode (explicit|semantic|crew|default)
)

// IncAgent counts one dispatch to the named agent (direct, pipeline, or crew).
func IncAgent(codename string) {
	if codename == "" {
		codename = "unknown"
	}
	requestsByAgent.Inc(codename)
}

// IncRoutingMode counts one routing decision by mode.
func IncRoutingMode(mode string) { routingMode.Inc(mode) }

// statusWriter captures the response status for error counting while
// preserving streaming (Flush) support.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Middleware counts every request, tracks in-flight, and records 5xx errors.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestsTotal.Add(1)
		InFlight.Add(1)
		defer InFlight.Add(-1)
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if sw.status >= 500 {
			ErrorsTotal.Add(1)
		}
	})
}

// Handler renders the Prometheus text exposition for GET /metrics.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder

	writeScalar := func(name, help, typ string, v int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s %s\n%s %d\n", name, help, name, typ, name, v)
	}
	writeScalar("eac_requests_total", "Total HTTP requests handled.", "counter", RequestsTotal.Load())
	writeScalar("eac_errors_total", "HTTP responses with a 5xx status.", "counter", ErrorsTotal.Load())
	writeScalar("eac_in_flight_requests", "In-flight HTTP requests.", "gauge", InFlight.Load())
	writeScalar("eac_upstream_answers_total", "Answers served by the upstream Copilot API.", "counter", UpstreamTotal.Load())
	writeScalar("eac_ollama_fallback_total", "Answers served by the local Ollama fallback.", "counter", OllamaFallbackTotal.Load())

	writeLabeled := func(name, help, label string, vals map[string]int64) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n", name, help, name)
		keys := make([]string, 0, len(vals))
		for k := range vals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "%s{%s=\"%s\"} %d\n", name, label, escapeLabel(k), vals[k])
		}
	}
	writeLabeled("eac_requests_by_agent_total", "Agent dispatches by codename.", "agent", requestsByAgent.snapshot())
	writeLabeled("eac_routing_mode_total", "Routing decisions by mode.", "mode", routingMode.snapshot())

	_, _ = w.Write([]byte(b.String()))
}

// escapeLabel escapes a Prometheus label value.
func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
