package handler_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestVersion_Public(t *testing.T) {
	srv := newTestServer()

	// Public — no token required.
	rec := do(t, srv, http.MethodGet, "/version", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("/version = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	var body struct {
		Version string `json:"version"`
		Go      string `json:"go"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Version != "v-test" {
		t.Errorf("version = %q, want the injected v-test", body.Version)
	}
	if !strings.HasPrefix(body.Go, "go") {
		t.Errorf("go = %q, want a runtime version", body.Go)
	}
}
