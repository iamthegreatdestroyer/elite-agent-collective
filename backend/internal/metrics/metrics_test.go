package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareAndHandler(t *testing.T) {
	before := RequestsTotal.Load()
	beforeErr := ErrorsTotal.Load()

	ok := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	ok.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/x", nil))

	bad := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	bad.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/y", nil))

	if RequestsTotal.Load() != before+2 {
		t.Fatalf("RequestsTotal: want +2 from %d, got %d", before, RequestsTotal.Load())
	}
	if ErrorsTotal.Load() != beforeErr+1 {
		t.Fatalf("ErrorsTotal: want +1 from %d, got %d", beforeErr, ErrorsTotal.Load())
	}
	if InFlight.Load() != 0 {
		t.Fatalf("InFlight should return to 0, got %d", InFlight.Load())
	}

	IncAgent("VELOCITY")
	IncRoutingMode("explicit")

	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest("GET", "/metrics", nil))
	body := rec.Body.String()
	for _, want := range []string{
		"# TYPE eac_requests_total counter",
		"# TYPE eac_in_flight_requests gauge",
		`eac_requests_by_agent_total{agent="VELOCITY"}`,
		`eac_routing_mode_total{mode="explicit"}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics output missing %q\n---\n%s", want, body)
		}
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Fatalf("bad content-type: %q", ct)
	}
}
