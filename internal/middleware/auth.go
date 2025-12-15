package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/service"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserClaimsKey is the context key for storing user claims
	UserClaimsKey contextKey = "user_claims"
)

// AuthMiddleware validates JWT tokens from requests
type AuthMiddleware struct {
	userService *service.UserService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := extractToken(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		// Validate token
		claims, err := m.userService.ValidateJWT(token)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Store claims in request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is a middleware that checks if user has required role
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(UserClaimsKey).(*service.JWTClaims)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "unauthorized: no claims found")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondWithError(w, http.StatusForbidden, "forbidden: insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth is a middleware that validates JWT if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token
		token, err := extractToken(r)
		if err != nil {
			// No token present, continue without claims
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		claims, err := m.userService.ValidateJWT(token)
		if err != nil {
			// Invalid token, continue without claims
			next.ServeHTTP(w, r)
			return
		}

		// Store claims in request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// GetUserClaims retrieves user claims from request context
func GetUserClaims(ctx context.Context) (*service.JWTClaims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(*service.JWTClaims)
	if !ok {
		return nil, errors.New("no user claims found in context")
	}
	return claims, nil
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
