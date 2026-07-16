package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeVerificationRepo is an in-memory domain.EmailVerificationRepository for
// handler tests.
type fakeVerificationRepo struct {
	byHash map[string]*domain.EmailVerificationToken
	nextID int
}

func newFakeVerificationRepo() *fakeVerificationRepo {
	return &fakeVerificationRepo{byHash: map[string]*domain.EmailVerificationToken{}, nextID: 1}
}

func (f *fakeVerificationRepo) Create(_ context.Context, t *domain.EmailVerificationToken) error {
	t.ID = f.nextID
	f.nextID++
	f.byHash[t.TokenHash] = t
	return nil
}
func (f *fakeVerificationRepo) FindByHash(_ context.Context, hash string) (*domain.EmailVerificationToken, error) {
	if t, ok := f.byHash[hash]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, domain.ErrVerificationTokenNotFound
}
func (f *fakeVerificationRepo) MarkUsed(_ context.Context, id int) error {
	for _, t := range f.byHash {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
			return nil
		}
	}
	return domain.ErrVerificationTokenNotFound
}

func tokenFromBody(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, "token=")
	if i < 0 {
		t.Fatalf("no token in email body: %q", body)
	}
	rest := body[i+len("token="):]
	return strings.FieldsFunc(rest, func(r rune) bool { return r == '\n' || r == '\r' || r == ' ' })[0]
}

func TestVerifyEmailEndpoint(t *testing.T) {
	srv, mail := newTestServerWithMailer()

	// Registering sends a verification email.
	do(t, srv, http.MethodPost, "/api/v2/auth/register", "",
		`{"username":"eve","email":"eve@example.com","password":"secret123"}`)
	msg, ok := mail.last()
	if !ok {
		t.Fatal("expected a verification email on registration")
	}
	token := tokenFromBody(t, msg.Body)

	// Verifying is public and succeeds.
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/verify-email", "",
		`{"token":"`+token+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify-email = %d (%s)", rec.Code, rec.Body)
	}

	// /me now reports the account as verified.
	login := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"eve@example.com","password":"secret123"}`)
	var lr struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body.Bytes(), &lr)
	me := do(t, srv, http.MethodGet, "/api/v2/auth/me", lr.Token, "")
	var mr struct {
		IsVerified bool `json:"is_verified"`
	}
	_ = json.Unmarshal(me.Body.Bytes(), &mr)
	if !mr.IsVerified {
		t.Fatalf("expected is_verified=true after verification, body=%s", me.Body)
	}
}

func TestVerifyEmailEndpoint_BadToken(t *testing.T) {
	srv, _ := newTestServerWithMailer()
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/verify-email", "",
		`{"token":"garbage"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad token = %d (%s), want 400", rec.Code, rec.Body)
	}
}

func TestResendVerification_RequiresAuth(t *testing.T) {
	srv, _ := newTestServerWithMailer()
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/resend-verification", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated resend = %d, want 401", rec.Code)
	}
}
