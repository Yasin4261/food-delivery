package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/router"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeUserRepo is a minimal in-memory domain.UserRepository for HTTP tests.
type fakeUserRepo struct {
	users  map[int]*domain.User
	nextID int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[int]*domain.User{}, nextID: 1}
}

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	u.ID = f.nextID
	f.nextID++
	cp := *u
	f.users[u.ID] = &cp
	return nil
}
func (f *fakeUserRepo) FindByID(_ context.Context, id int) (*domain.User, error) {
	if u, ok := f.users[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, domain.ErrUserNotFound
}
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (f *fakeUserRepo) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	for _, u := range f.users {
		if u.Username == username {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

// healthHandler needs a Pinger; this one always reports healthy.
type okPinger struct{}

func (okPinger) PingContext(context.Context) error { return nil }

// newTestServer wires the real handler stack over in-memory fake repositories
// and returns the configured HTTP handler.
func newTestServer() http.Handler {
	chefRepo := newFakeChefRepo()
	itemRepo := newFakeMenuItemRepo()
	authService := service.NewAuthService(newFakeUserRepo(), "test-secret", time.Hour)
	chefService := service.NewChefService(chefRepo)
	menuService := service.NewMenuService(chefRepo, newFakeMenuRepo(), itemRepo)
	orderService := service.NewOrderService(newFakeOrderRepo(), itemRepo, chefRepo)
	authMiddleware := middleware.NewAuth(authService)
	healthHandler := handler.NewHealthHandler(okPinger{})
	authHandler := handler.NewAuthHandler(authService)
	chefHandler := handler.NewChefHandler(chefService)
	menuHandler := handler.NewMenuHandler(menuService)
	orderHandler := handler.NewOrderHandler(orderService)
	return router.NewRouter(authMiddleware, healthHandler, authHandler, chefHandler, menuHandler, orderHandler).Setup()
}

// registerAndToken registers a user through the API and returns its bearer token.
func registerAndToken(t *testing.T, srv http.Handler, username, email string) string {
	t.Helper()
	body := `{"username":"` + username + `","email":"` + email + `","password":"secret123","role":"chef"}`
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/register", "", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup register failed: %d (%s)", rec.Code, rec.Body)
	}
	var reg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &reg); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	return reg.Token
}

func do(t *testing.T, srv http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestAuthFlow_HTTP(t *testing.T) {
	srv := newTestServer()

	// Register.
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/register", "",
		`{"username":"yasin","email":"yasin@example.com","password":"secret123","role":"chef"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201 (body: %s)", rec.Code, rec.Body)
	}
	var reg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &reg); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if reg.Token == "" {
		t.Fatal("expected a token in register response")
	}

	// /me without a token -> 401.
	if rec := do(t, srv, http.MethodGet, "/api/v2/auth/me", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("/me without token = %d, want 401", rec.Code)
	}

	// /me with the token -> 200 and the right user.
	rec = do(t, srv, http.MethodGet, "/api/v2/auth/me", reg.Token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("/me with token = %d, want 200 (body: %s)", rec.Code, rec.Body)
	}
	var me domain.User
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatalf("decode /me response: %v", err)
	}
	if me.Email != "yasin@example.com" || me.Role != domain.RoleChef {
		t.Errorf("/me returned %+v, want yasin@example.com / chef", me)
	}
	if me.PasswordHash != "" {
		t.Error("/me must not expose the password hash")
	}
}

func TestAuthHTTP_ErrorCodes(t *testing.T) {
	srv := newTestServer()
	_ = do(t, srv, http.MethodPost, "/api/v2/auth/register", "",
		`{"username":"yasin","email":"yasin@example.com","password":"secret123"}`)

	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"duplicate email", http.MethodPost, "/api/v2/auth/register",
			`{"username":"other","email":"yasin@example.com","password":"secret123"}`, http.StatusConflict},
		{"validation", http.MethodPost, "/api/v2/auth/register",
			`{"username":"ab","email":"a@b.c","password":"123"}`, http.StatusBadRequest},
		{"bad json", http.MethodPost, "/api/v2/auth/login", `{not json`, http.StatusBadRequest},
		{"wrong password", http.MethodPost, "/api/v2/auth/login",
			`{"email":"yasin@example.com","password":"nope"}`, http.StatusUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(t, srv, tc.method, tc.path, "", tc.body); rec.Code != tc.want {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tc.want, rec.Body)
			}
		})
	}
}
