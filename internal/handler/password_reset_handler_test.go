package handler_test

import (
	"net/http"
	"strings"
	"testing"
)

// resetTokenFromEmail extracts the token from a reset email body containing
// "...?token=<raw>".
func resetTokenFromEmail(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, "token=")
	if i < 0 {
		t.Fatalf("no token in email body: %q", body)
	}
	return strings.FieldsFunc(body[i+len("token="):], func(r rune) bool {
		return r == '\n' || r == '\r' || r == ' '
	})[0]
}

func TestPasswordReset_HTTPFlow(t *testing.T) {
	srv, mail := newTestServerWithMailer()
	_ = registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Forgot-password always 202, and never returns the token in the body.
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/forgot-password", "", `{"email":"cust@example.com"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("forgot-password = %d, want 202 (%s)", rec.Code, rec.Body)
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "token") {
		t.Errorf("response must not contain the reset token: %s", rec.Body)
	}

	// The token arrives by email instead.
	sent, ok := mail.last()
	if !ok || sent.To != "cust@example.com" {
		t.Fatalf("expected a reset email, got %+v", sent)
	}
	token := resetTokenFromEmail(t, sent.Body)

	// Reset with the token succeeds and the new password works.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"`+token+`","password":"brandnew"}`); rec.Code != http.StatusOK {
		t.Fatalf("reset-password = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"cust@example.com","password":"brandnew"}`); rec.Code != http.StatusOK {
		t.Errorf("login with new password = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	// The token is single-use → second attempt is 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/reset-password", "",
		`{"token":"`+token+`","password":"brandnew2"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("reused token = %d, want 400", rec.Code)
	}
}

func TestPasswordReset_HTTPGuards(t *testing.T) {
	srv, mail := newTestServerWithMailer()

	// Unknown email still returns 202 (no account enumeration) and sends nothing.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/forgot-password", "", `{"email":"ghost@example.com"}`); rec.Code != http.StatusAccepted {
		t.Errorf("forgot unknown = %d, want 202", rec.Code)
	}
	if _, ok := mail.last(); ok {
		t.Error("no email should be sent for an unknown address")
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
