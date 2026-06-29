package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPasswordReset_HTTPFlow(t *testing.T) {
	srv := newTestServer()
	// Register a user (registerCustomerToken registers cust@example.com).
	_ = registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Forgot-password always 202; in test mode the token is echoed back.
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/forgot-password", "", `{"email":"cust@example.com"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("forgot-password = %d, want 202 (%s)", rec.Code, rec.Body)
	}
	var resp struct {
		ResetToken string `json:"reset_token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.ResetToken == "" {
		t.Fatal("expected a reset_token in dev mode")
	}

	// Reset with the token succeeds and the new password works.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"`+resp.ResetToken+`","password":"brandnew"}`); rec.Code != http.StatusOK {
		t.Fatalf("reset-password = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"cust@example.com","password":"brandnew"}`); rec.Code != http.StatusOK {
		t.Errorf("login with new password = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	// The token is single-use → second attempt is 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"`+resp.ResetToken+`","password":"brandnew2"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("reused token = %d, want 400", rec.Code)
	}
}

func TestPasswordReset_HTTPGuards(t *testing.T) {
	srv := newTestServer()

	// Unknown email still returns 202 (no account enumeration) and no token.
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/forgot-password", "", `{"email":"ghost@example.com"}`)
	if rec.Code != http.StatusAccepted {
		t.Errorf("forgot unknown = %d, want 202", rec.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["reset_token"]; ok {
		t.Error("no token should be issued for an unknown email")
	}

	// Garbage token → 400; short password → 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"nope","password":"longenough"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("garbage token = %d, want 400", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"nope","password":"123"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("short password = %d, want 400", rec.Code)
	}
}
