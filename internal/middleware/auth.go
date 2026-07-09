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

// Denylist reports whether a token id (jti) has been revoked.
type Denylist interface {
	IsRevoked(jti string) bool
}

type contextKey string

const claimsKey contextKey = "auth_claims"

// Auth holds the dependencies for authentication middleware.
type Auth struct {
	parser   TokenParser
	denylist Denylist
}

// NewAuth builds the auth middleware. denylist may be nil to disable revocation
// checks.
func NewAuth(parser TokenParser, denylist Denylist) *Auth {
	return &Auth{parser: parser, denylist: denylist}
}

// bearerToken extracts the caller's token. Normally it comes from the
// "Authorization: Bearer <token>" header. For WebSocket upgrade requests only,
// an `access_token` query parameter is accepted as a fallback, because the
// browser WebSocket API cannot set headers. The request logger records only
// the path (never the query), so the token does not reach logs.
func bearerToken(r *http.Request) string {
	if header := r.Header.Get("Authorization"); header != "" {
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			return ""
		}
		return strings.TrimSpace(token)
	}
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return strings.TrimSpace(r.URL.Query().Get("access_token"))
	}
	return ""
}

// Require rejects requests without a valid bearer token (see bearerToken for
// where it may come from). On success it stores the claims in the request
// context.
func (a *Auth) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing or malformed bearer token")
			return
		}

		claims, err := a.parser.ParseToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if a.denylist != nil && a.denylist.IsRevoked(claims.ID) {
			writeError(w, http.StatusUnauthorized, "token has been revoked")
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
