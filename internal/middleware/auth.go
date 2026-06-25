package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/service"
)

// TokenParser is the slice of AuthService the middleware needs. Declaring it as
// an interface keeps the middleware decoupled from the concrete service.
type TokenParser interface {
	ParseToken(token string) (*service.Claims, error)
}

type contextKey string

const claimsKey contextKey = "auth_claims"

// Auth holds the dependencies for authentication middleware.
type Auth struct {
	parser TokenParser
}

// NewAuth builds the auth middleware.
func NewAuth(parser TokenParser) *Auth {
	return &Auth{parser: parser}
}

// Require rejects requests without a valid "Authorization: Bearer <token>"
// header. On success it stores the claims in the request context.
func (a *Auth) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			writeError(w, http.StatusUnauthorized, "missing or malformed bearer token")
			return
		}

		claims, err := a.parser.ParseToken(strings.TrimSpace(token))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole wraps Require and additionally enforces that the caller has one
// of the allowed roles.
func (a *Auth) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return a.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "insufficient permissions")
		}))
	}
}

// ClaimsFromContext returns the authenticated claims stored by Require.
func ClaimsFromContext(ctx context.Context) (*service.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*service.Claims)
	return claims, ok
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
