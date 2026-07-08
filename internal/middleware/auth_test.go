package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeParser returns canned claims or an error.
type fakeParser struct {
	claims *service.Claims
	err    error
}

func (f fakeParser) ParseToken(string) (*service.Claims, error) { return f.claims, f.err }

// fakeDenylist marks a single jti as revoked.
type fakeDenylist struct{ revokedJTI string }

func (f fakeDenylist) IsRevoked(jti string) bool { return jti != "" && jti == f.revokedJTI }

func chefClaims(jti string) *service.Claims {
	return &service.Claims{
		UserID: 7,
		Role:   "chef",
		RegisteredClaims: jwt.RegisteredClaims{
			ID: jti,
		},
	}
}

// echo records whether the wrapped handler ran and what claims it saw.
func echo(t *testing.T, sawClaims *bool) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		*sawClaims = ok && claims != nil
		w.WriteHeader(http.StatusOK)
	})
}

func doAuth(h http.Handler, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRequire_TokenHandling(t *testing.T) {
	var saw bool
	ok := middleware.NewAuth(fakeParser{claims: chefClaims("jti-1")}, fakeDenylist{})
	bad := middleware.NewAuth(fakeParser{err: errors.New("bad token")}, fakeDenylist{})

	cases := []struct {
		name string
		auth *middleware.Auth
		hdr  string
		want int
	}{
		{"missing header", ok, "", http.StatusUnauthorized},
		{"not bearer", ok, "Basic abc", http.StatusUnauthorized},
		{"empty bearer", ok, "Bearer   ", http.StatusUnauthorized},
		{"parser rejects", bad, "Bearer whatever", http.StatusUnauthorized},
		{"valid", ok, "Bearer good-token", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			saw = false
			rec := doAuth(tc.auth.Require(echo(t, &saw)), tc.hdr)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.want == http.StatusOK && !saw {
				t.Error("handler should have run with claims in context")
			}
			if tc.want != http.StatusOK && saw {
				t.Error("handler must not run on rejected requests")
			}
		})
	}
}

func TestRequire_RevokedTokenRejected(t *testing.T) {
	var saw bool
	auth := middleware.NewAuth(fakeParser{claims: chefClaims("revoked-jti")}, fakeDenylist{revokedJTI: "revoked-jti"})

	rec := doAuth(auth.Require(echo(t, &saw)), "Bearer stolen")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("revoked token = %d, want 401", rec.Code)
	}
	if saw {
		t.Error("handler must not run for a revoked token")
	}
}

func TestRequire_NilDenylistAllowed(t *testing.T) {
	var saw bool
	auth := middleware.NewAuth(fakeParser{claims: chefClaims("jti-1")}, nil)

	if rec := doAuth(auth.Require(echo(t, &saw)), "Bearer tok"); rec.Code != http.StatusOK || !saw {
		t.Errorf("nil denylist should disable revocation checks, got %d", rec.Code)
	}
}

func TestRequireRole(t *testing.T) {
	var saw bool
	auth := middleware.NewAuth(fakeParser{claims: chefClaims("jti-1")}, fakeDenylist{})

	// Matching role passes.
	if rec := doAuth(auth.RequireRole("chef")(echo(t, &saw)), "Bearer tok"); rec.Code != http.StatusOK {
		t.Errorf("chef role = %d, want 200", rec.Code)
	}
	// Any of several roles passes.
	if rec := doAuth(auth.RequireRole("admin", "chef")(echo(t, &saw)), "Bearer tok"); rec.Code != http.StatusOK {
		t.Errorf("multi-role = %d, want 200", rec.Code)
	}
	// Wrong role -> 403.
	saw = false
	if rec := doAuth(auth.RequireRole("admin")(echo(t, &saw)), "Bearer tok"); rec.Code != http.StatusForbidden || saw {
		t.Errorf("wrong role = %d (handler ran: %v), want 403 without running", rec.Code, saw)
	}
}
