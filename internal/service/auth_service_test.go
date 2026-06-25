package service_test

import (
	"context"
	"errors"
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

func newService(repo domain.UserRepository) *service.AuthService {
	return service.NewAuthService(repo, "test-secret", time.Hour)
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

func TestParseToken_Rejected(t *testing.T) {
	svc := newService(newFakeUserRepo())

	if _, err := svc.ParseToken("not-a-real-token"); err == nil {
		t.Error("expected an error for a garbage token")
	}

	// A token signed with a different secret must be rejected.
	other := service.NewAuthService(newFakeUserRepo(), "different-secret", time.Hour)
	res, _ := other.Register(context.Background(), validRegister())
	if _, err := svc.ParseToken(res.Token); err == nil {
		t.Error("expected rejection of a token signed with another secret")
	}
}
