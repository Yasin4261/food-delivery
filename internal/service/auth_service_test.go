package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeUserRepo is an in-memory domain.UserRepository for tests. Because the
// service depends on the interface (a port), no database is needed here — this
// is the payoff of the hexagonal layering.
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
	return nil
}

// fakeResetRepo is an in-memory domain.PasswordResetRepository for tests.
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

// recordingMailer is a domain.Mailer that captures sent emails for assertions.
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

// extractToken pulls the reset token out of an email body containing
// "...?token=<raw>".
func extractToken(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, "token=")
	if i < 0 {
		t.Fatalf("no token in email body: %q", body)
	}
	rest := body[i+len("token="):]
	return strings.FieldsFunc(rest, func(r rune) bool { return r == '\n' || r == '\r' || r == ' ' })[0]
}

func newService(repo domain.UserRepository) *service.AuthService {
	return service.NewAuthService(repo, newFakeResetRepo(), &recordingMailer{}, "test-secret", time.Hour, "http://app.test")
}

// newServiceWithResets builds an AuthService over the given fakes so reset tests
// can reach into the token store and the sent mail.
func newServiceWithResets(repo domain.UserRepository, resets domain.PasswordResetRepository, mail domain.Mailer) *service.AuthService {
	return service.NewAuthService(repo, resets, mail, "test-secret", time.Hour, "http://app.test")
}

func validRegister() service.RegisterInput {
	return service.RegisterInput{
		Username: "yasin",
		Email:    "Yasin@Example.com",
		Password: "secret123",
	}
}

func TestRegister_Success(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newService(repo)

	res, err := svc.Register(context.Background(), validRegister())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" {
		t.Error("expected a token")
	}
	if res.User.ID == 0 {
		t.Error("expected an assigned ID")
	}
	if res.User.Role != domain.RoleCustomer {
		t.Errorf("default role = %q, want customer", res.User.Role)
	}
	if res.User.Email != "yasin@example.com" {
		t.Errorf("email should be normalised to lower case, got %q", res.User.Email)
	}
	if res.User.PasswordHash != "" {
		t.Error("password hash must not be exposed in the result")
	}

	// The stored hash must verify against the original password.
	stored, _ := repo.FindByEmail(context.Background(), "yasin@example.com")
	if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("secret123")) != nil {
		t.Error("stored password hash does not match the original password")
	}
}

func TestRegister_CustomRole(t *testing.T) {
	svc := newService(newFakeUserRepo())
	in := validRegister()
	in.Role = domain.RoleChef

	res, err := svc.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.User.Role != domain.RoleChef {
		t.Errorf("role = %q, want chef", res.User.Role)
	}
}

func TestRegister_AdminCannotSelfAssign(t *testing.T) {
	svc := newService(newFakeUserRepo())
	in := validRegister()
	in.Role = domain.RoleAdmin

	_, err := svc.Register(context.Background(), in)
	var ve service.ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("err = %v, want ValidationError (admin must not be self-assignable)", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newService(newFakeUserRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, validRegister()); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	in := validRegister()
	in.Username = "different"
	_, err := svc.Register(ctx, in)
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Errorf("err = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc := newService(newFakeUserRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, validRegister()); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	in := validRegister()
	in.Email = "other@example.com"
	_, err := svc.Register(ctx, in)
	if !errors.Is(err, domain.ErrUsernameAlreadyExists) {
		t.Errorf("err = %v, want ErrUsernameAlreadyExists", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	svc := newService(newFakeUserRepo())
	cases := map[string]func(*service.RegisterInput){
		"short username": func(in *service.RegisterInput) { in.Username = "ab" },
		"bad email":      func(in *service.RegisterInput) { in.Email = "no-at-sign" },
		"short password": func(in *service.RegisterInput) { in.Password = "123" },
		"invalid role":   func(in *service.RegisterInput) { in.Role = "wizard" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := validRegister()
			mutate(&in)
			_, err := svc.Register(context.Background(), in)
			var ve service.ValidationError
			if !errors.As(err, &ve) {
				t.Errorf("err = %v, want ValidationError", err)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newService(newFakeUserRepo())
	ctx := context.Background()
	if _, err := svc.Register(ctx, validRegister()); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	res, err := svc.Login(ctx, "yasin@example.com", "secret123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if res.Token == "" {
		t.Error("expected a token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newService(newFakeUserRepo())
	ctx := context.Background()
	_, _ = svc.Register(ctx, validRegister())

	_, err := svc.Login(ctx, "yasin@example.com", "wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnknownEmailLooksLikeBadCredentials(t *testing.T) {
	svc := newService(newFakeUserRepo())

	_, err := svc.Login(context.Background(), "ghost@example.com", "whatever")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials (must not leak ErrUserNotFound)", err)
	}
}

func TestLogin_InactiveAccount(t *testing.T) {
	repo := newFakeUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	_ = repo.Create(context.Background(), &domain.User{
		Username:     "frozen",
		Email:        "frozen@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleCustomer,
		IsActive:     false,
	})
	svc := newService(repo)

	_, err := svc.Login(context.Background(), "frozen@example.com", "secret123")
	if !errors.Is(err, domain.ErrAccountInactive) {
		t.Errorf("err = %v, want ErrAccountInactive", err)
	}
}

func TestParseToken_RoundTrip(t *testing.T) {
	svc := newService(newFakeUserRepo())
	res, err := svc.Register(context.Background(), validRegister())
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	claims, err := svc.ParseToken(res.Token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != res.User.ID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, res.User.ID)
	}
	if claims.Role != domain.RoleCustomer {
		t.Errorf("claims.Role = %q, want customer", claims.Role)
	}
}

func TestPasswordReset_Flow(t *testing.T) {
	repo := newFakeUserRepo()
	resets := newFakeResetRepo()
	mail := &recordingMailer{}
	svc := newServiceWithResets(repo, resets, mail)
	ctx := context.Background()
	if _, err := svc.Register(ctx, validRegister()); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Unknown email is a silent no-op: no error, and no email sent.
	if err := svc.RequestPasswordReset(ctx, "ghost@example.com"); err != nil {
		t.Errorf("unknown email = %v, want nil", err)
	}
	if _, ok := mail.last(); ok {
		t.Error("no email should be sent for an unknown address")
	}

	// Known email: an email is sent carrying the reset link.
	if err := svc.RequestPasswordReset(ctx, "yasin@example.com"); err != nil {
		t.Fatalf("request reset: %v", err)
	}
	sent, ok := mail.last()
	if !ok || sent.To != "yasin@example.com" {
		t.Fatalf("expected a reset email to the user, got %+v", sent)
	}
	token := extractToken(t, sent.Body)

	if err := svc.ResetPassword(ctx, token, "newsecret"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	// New password works; old one no longer does.
	if _, err := svc.Login(ctx, "yasin@example.com", "newsecret"); err != nil {
		t.Errorf("login with new password failed: %v", err)
	}
	if _, err := svc.Login(ctx, "yasin@example.com", "secret123"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("old password still works: %v", err)
	}

	// Single-use: the same token cannot be redeemed twice.
	if err := svc.ResetPassword(ctx, token, "another1"); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("reused token = %v, want ErrInvalidResetToken", err)
	}
}

func TestPasswordReset_RejectsBadTokens(t *testing.T) {
	repo := newFakeUserRepo()
	resets := newFakeResetRepo()
	mail := &recordingMailer{}
	svc := newServiceWithResets(repo, resets, mail)
	ctx := context.Background()
	_, _ = svc.Register(ctx, validRegister())

	// Unknown token.
	if err := svc.ResetPassword(ctx, "garbage", "newsecret"); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("unknown token = %v, want ErrInvalidResetToken", err)
	}

	// Expired token.
	if err := svc.RequestPasswordReset(ctx, "yasin@example.com"); err != nil {
		t.Fatalf("request: %v", err)
	}
	sent, _ := mail.last()
	token := extractToken(t, sent.Body)
	for _, stored := range resets.byHash {
		stored.ExpiresAt = time.Now().Add(-time.Minute)
	}
	if err := svc.ResetPassword(ctx, token, "newsecret"); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("expired token = %v, want ErrInvalidResetToken", err)
	}

	// Short password is a validation error (checked before the token lookup).
	var ve service.ValidationError
	if err := svc.ResetPassword(ctx, "anything", "123"); !errors.As(err, &ve) {
		t.Errorf("short password = %v, want ValidationError", err)
	}
}

func TestParseToken_Rejected(t *testing.T) {
	svc := newService(newFakeUserRepo())

	if _, err := svc.ParseToken("not-a-real-token"); err == nil {
		t.Error("expected an error for a garbage token")
	}

	// A token signed with a different secret must be rejected.
	other := service.NewAuthService(newFakeUserRepo(), newFakeResetRepo(), &recordingMailer{}, "different-secret", time.Hour, "http://app.test")
	res, _ := other.Register(context.Background(), validRegister())
	if _, err := svc.ParseToken(res.Token); err == nil {
		t.Error("expected rejection of a token signed with another secret")
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newService(repo)
	ctx := context.Background()
	reg, err := svc.Register(ctx, validRegister())
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	uid := reg.User.ID

	t.Run("wrong current password", func(t *testing.T) {
		err := svc.ChangePassword(ctx, uid, "not-the-password", "newpass1")
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("err = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("too-short new password", func(t *testing.T) {
		if err := svc.ChangePassword(ctx, uid, "secret123", "abc"); !isValidation(err) {
			t.Errorf("err = %v, want ValidationError", err)
		}
	})

	t.Run("success rotates the hash", func(t *testing.T) {
		if err := svc.ChangePassword(ctx, uid, "secret123", "newpass1"); err != nil {
			t.Fatalf("change: %v", err)
		}
		if _, err := svc.Login(ctx, "yasin@example.com", "secret123"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("old password still works: %v", err)
		}
		if _, err := svc.Login(ctx, "yasin@example.com", "newpass1"); err != nil {
			t.Errorf("new password rejected: %v", err)
		}
	})
}

func TestAuthService_UpdateProfile(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newService(repo)
	ctx := context.Background()
	reg, err := svc.Register(ctx, validRegister())
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	uid := reg.User.ID

	t.Run("lat without lng rejected", func(t *testing.T) {
		lat := 41.0
		_, err := svc.UpdateProfile(ctx, uid, service.UpdateProfileInput{Latitude: &lat})
		if !isValidation(err) {
			t.Errorf("err = %v, want ValidationError", err)
		}
	})

	t.Run("updates contact and location, never identity", func(t *testing.T) {
		lat, lng := 41.0082, 28.9784
		got, err := svc.UpdateProfile(ctx, uid, service.UpdateProfileInput{
			PhoneNumber: "+90 555 000 00 00",
			Address:     "New Street 5",
			City:        "Istanbul",
			Latitude:    &lat,
			Longitude:   &lng,
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if got.PhoneNumber == nil || *got.PhoneNumber != "+90 555 000 00 00" || got.City == nil || *got.City != "Istanbul" {
			t.Errorf("profile not applied: %+v", got)
		}
		if got.Email != "yasin@example.com" || got.Username != "yasin" || got.Role != domain.RoleCustomer {
			t.Errorf("identity fields must be untouched: %+v", got)
		}
		if got.PasswordHash != "" {
			t.Error("password hash leaked in response")
		}
		stored, _ := repo.FindByID(ctx, uid)
		if stored.Latitude == nil || *stored.Latitude != lat {
			t.Errorf("location not persisted: %+v", stored)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		if _, err := svc.UpdateProfile(ctx, 999, service.UpdateProfileInput{}); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("err = %v, want ErrUserNotFound", err)
		}
	})
}
