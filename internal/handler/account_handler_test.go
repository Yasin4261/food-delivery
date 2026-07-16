package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeAccountRepo is an in-memory domain.AccountRepository that anonymises the
// user in the shared fakeUserRepo, so a follow-up login/token check behaves
// like the real deactivation.
type fakeAccountRepo struct{ users *fakeUserRepo }

func newFakeAccountRepo(u *fakeUserRepo) *fakeAccountRepo { return &fakeAccountRepo{users: u} }

func (f *fakeAccountRepo) Anonymise(_ context.Context, userID int) error {
	u, ok := f.users.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.IsActive = false
	u.Email = fmt.Sprintf("deleted-%d@removed.invalid", userID)
	u.Username = fmt.Sprintf("deleted-%d", userID)
	u.PasswordHash = ""
	return nil
}

func TestAccount_ExportOwnData(t *testing.T) {
	srv := newTestServer()
	customer := registerCustomerToken(t, srv, "expuser", "exp@example.com")

	// Anonymous is rejected.
	if rec := do(t, srv, http.MethodGet, "/api/v2/users/me/export", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon export = %d, want 401", rec.Code)
	}

	rec := do(t, srv, http.MethodGet, "/api/v2/users/me/export", customer, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export = %d (%s)", rec.Code, rec.Body)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Errorf("missing attachment disposition, got %q", cd)
	}
	// The password hash must never appear in the dump.
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Error("export leaked password_hash")
	}
	var export domain.AccountExport
	if err := json.Unmarshal(rec.Body.Bytes(), &export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if export.User == nil || export.User.Email != "exp@example.com" {
		t.Errorf("export user wrong: %+v", export.User)
	}
	// Collections are always present (never null) even when empty.
	if export.Addresses == nil || export.Orders == nil || export.Reviews == nil || export.Conversations == nil {
		t.Errorf("export collections should be non-nil: %+v", export)
	}
}

func TestAccount_DeleteRequiresPasswordAndRevokes(t *testing.T) {
	srv := newTestServer()
	customer := registerCustomerToken(t, srv, "deluser", "del@example.com")

	// Wrong password is rejected; the account stays usable.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/users/me", customer, `{"password":"wrongpass"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("delete with wrong password = %d, want 401 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodGet, "/api/v2/auth/me", customer, ""); rec.Code != http.StatusOK {
		t.Fatalf("account should still work after a failed delete = %d", rec.Code)
	}

	// Correct password deletes the account.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/users/me", customer, `{"password":"secret123"}`); rec.Code != http.StatusOK {
		t.Fatalf("delete = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	// The presented token is revoked immediately.
	if rec := do(t, srv, http.MethodGet, "/api/v2/auth/me", customer, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("token should be revoked after delete = %d, want 401", rec.Code)
	}
	// Login with the old credentials fails (account deactivated + email scrubbed).
	login := do(t, srv, http.MethodPost, "/api/v2/auth/login", "", `{"email":"del@example.com","password":"secret123"}`)
	if login.Code != http.StatusUnauthorized {
		t.Errorf("login after delete = %d, want 401", login.Code)
	}
}
