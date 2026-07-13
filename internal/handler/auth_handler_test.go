package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/payment"
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
func (f *fakeUserRepo) UpdatePassword(_ context.Context, userID int, passwordHash string) error {
	if u, ok := f.users[userID]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return domain.ErrUserNotFound
}

func (f *fakeUserRepo) UpdateProfile(_ context.Context, u *domain.User) error {
	stored, ok := f.users[u.ID]
	if !ok {
		return domain.ErrUserNotFound
	}
	stored.PhoneNumber = u.PhoneNumber
	stored.Address = u.Address
	stored.City = u.City
	stored.State = u.State
	stored.ZipCode = u.ZipCode
	stored.Latitude = u.Latitude
	stored.Longitude = u.Longitude
	stored.EmailNotifications = u.EmailNotifications
	return nil
}

// fakeResetRepo is an in-memory domain.PasswordResetRepository for HTTP tests.
type fakeResetRepo struct {
	byHash map[string]*domain.PasswordResetToken
	nextID int
}

func newFakeResetRepo() *fakeResetRepo {
	return &fakeResetRepo{byHash: map[string]*domain.PasswordResetToken{}, nextID: 1}
}

func (f *fakeResetRepo) Create(_ context.Context, t *domain.PasswordResetToken) error {
	t.ID = f.nextID
	f.nextID++
	f.byHash[t.TokenHash] = t
	return nil
}
func (f *fakeResetRepo) FindByHash(_ context.Context, hash string) (*domain.PasswordResetToken, error) {
	if t, ok := f.byHash[hash]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, domain.ErrResetTokenNotFound
}
func (f *fakeResetRepo) MarkUsed(_ context.Context, id int) error {
	for _, t := range f.byHash {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
			return nil
		}
	}
	return domain.ErrResetTokenNotFound
}

// healthHandler needs a Pinger; this one always reports healthy.
type okPinger struct{}

func (okPinger) PingContext(context.Context) error { return nil }

// newTestServer wires the real handler stack over in-memory fake repositories
// and returns the configured HTTP handler.
func newTestServer() http.Handler {
	srv, _ := newTestServerWithMailer()
	return srv
}

// recordingMailer captures emails sent during HTTP tests (e.g. the reset link).
type recordingMailer struct {
	mu   sync.Mutex
	sent []domain.Email
}

func (m *recordingMailer) Send(_ context.Context, msg domain.Email) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, msg)
	return nil
}
func (m *recordingMailer) last() (domain.Email, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sent) == 0 {
		return domain.Email{}, false
	}
	return m.sent[len(m.sent)-1], true
}

// newTestServerWithMailer is like newTestServer but exposes the mailer so reset
// tests can read the emitted reset link.
func newTestServerWithMailer() (http.Handler, *recordingMailer) {
	chefRepo := newFakeChefRepo()
	itemRepo := newFakeMenuItemRepo()
	userRepo := newFakeUserRepo()
	mail := &recordingMailer{}
	authService := service.NewAuthService(userRepo, newFakeResetRepo(), mail, "test-secret", time.Hour, "http://app.test")
	chefService := service.NewChefService(chefRepo)
	menuService := service.NewMenuService(chefRepo, newFakeMenuRepo(), itemRepo)
	orderRepo := newFakeOrderRepo()
	paymentService := service.NewPaymentService(newFakePaymentSessionRepo(), orderRepo, userRepo, payment.NewMock("http://app.test"), "http://app.test")
	orderService := service.NewOrderService(orderRepo, itemRepo, chefRepo, paymentService, nil)
	favoriteService := service.NewFavoriteService(newFakeFavoriteRepo(chefRepo), chefRepo)
	reviewService := service.NewReviewService(newFakeReviewRepo(), orderRepo)
	earningsService := service.NewEarningsService(newFakeEarningsRepo(), chefRepo)
	searchService := service.NewSearchService(newFakeSearchRepo())
	chatService := service.NewChatService(newFakeChatRepo(), chefRepo)
	denylist := service.NewTokenDenylist()
	authMiddleware := middleware.NewAuth(authService, denylist)
	healthHandler := handler.NewHealthHandler(okPinger{})
	authHandler := handler.NewAuthHandler(authService, denylist)
	chefHandler := handler.NewChefHandler(chefService)
	menuHandler := handler.NewMenuHandler(menuService)
	orderHandler := handler.NewOrderHandler(orderService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	earningsHandler := handler.NewEarningsHandler(earningsService)
	searchHandler := handler.NewSearchHandler(searchService)
	chatHandler := handler.NewChatHandler(chatService)
	versionHandler := handler.NewVersionHandler("v-test")
	paymentHandler := handler.NewPaymentHandler(paymentService)
	// A generous budget so no test trips the per-IP throttle accidentally.
	authLimiter := middleware.NewRateLimiter(1000, time.Minute)
	return router.NewRouter(authMiddleware, healthHandler, authHandler, chefHandler, menuHandler, orderHandler, favoriteHandler, reviewHandler, earningsHandler, searchHandler, chatHandler, versionHandler, paymentHandler, authLimiter).Setup(), mail
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

// pageResp mirrors the handler's list envelope { data, limit, offset, total }.
type pageResp[T any] struct {
	Data   []T `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// decodePage unmarshals a paginated list response.
func decodePage[T any](t *testing.T, body []byte) pageResp[T] {
	t.Helper()
	var p pageResp[T]
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("decode page: %v (%s)", err, body)
	}
	return p
}

// doForm sends an application/x-www-form-urlencoded POST (like a browser form
// or a payment-gateway redirect).
func doForm(t *testing.T, srv http.Handler, path, form string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
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

func TestLogout_RevokesToken(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")

	// The token works before logout.
	if rec := do(t, srv, http.MethodGet, "/api/v2/auth/me", token, ""); rec.Code != http.StatusOK {
		t.Fatalf("/me before logout = %d, want 200", rec.Code)
	}

	// Logout (authenticated) revokes this token.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/logout", token, ""); rec.Code != http.StatusOK {
		t.Fatalf("logout = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	// The same token is now rejected before its natural expiry.
	if rec := do(t, srv, http.MethodGet, "/api/v2/auth/me", token, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("/me after logout = %d, want 401 (revoked)", rec.Code)
	}

	// Logout itself now requires a valid token.
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/logout", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("logout without token = %d, want 401", rec.Code)
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
