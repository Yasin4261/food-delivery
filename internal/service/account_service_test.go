package service_test

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeAccountRepo records the anonymised user ids (or returns a preset error).
type fakeAccountRepo struct {
	anonymised []int
	err        error
}

func (f *fakeAccountRepo) Anonymise(_ context.Context, userID int) error {
	if f.err != nil {
		return f.err
	}
	f.anonymised = append(f.anonymised, userID)
	return nil
}

func newAccountService(users domain.UserRepository, chefs domain.ChefRepository, acct domain.AccountRepository) *service.AccountService {
	return service.NewAccountService(users, chefs, newFakeAddressRepo(), newFakeOrderRepo(), newFakeReviewRepo(), newFakeChatRepo(), acct)
}

func TestAccountService_Delete_PasswordGate(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost)
	u := &domain.User{Email: "a@b.c", PasswordHash: string(hash), IsActive: true}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	acct := &fakeAccountRepo{}
	svc := newAccountService(repo, newFakeChefRepo(), acct)

	// Wrong password: rejected, and nothing is anonymised.
	if err := svc.Delete(ctx, u.ID, "wrong"); err != domain.ErrInvalidCredentials {
		t.Fatalf("wrong password = %v, want ErrInvalidCredentials", err)
	}
	if len(acct.anonymised) != 0 {
		t.Fatalf("account must not be anonymised on a bad password: %v", acct.anonymised)
	}

	// Correct password: anonymises the right user.
	if err := svc.Delete(ctx, u.ID, "hunter2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(acct.anonymised) != 1 || acct.anonymised[0] != u.ID {
		t.Fatalf("anonymised = %v, want [%d]", acct.anonymised, u.ID)
	}
}

func TestAccountService_Export_IncludesChefAndClearsHash(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost)
	u := &domain.User{Email: "chef@b.c", PasswordHash: string(hash), Role: domain.RoleChef, IsActive: true}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	chefs := newFakeChefRepo()
	if err := chefs.Create(ctx, &domain.Chef{UserID: u.ID, BusinessName: "K", IsActive: true}); err != nil {
		t.Fatalf("seed chef: %v", err)
	}
	svc := newAccountService(repo, chefs, &fakeAccountRepo{})

	export, err := svc.Export(ctx, u.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if export.User == nil || export.User.PasswordHash != "" {
		t.Errorf("export must include the user with a cleared hash: %+v", export.User)
	}
	if export.Chef == nil || export.Chef.BusinessName != "K" {
		t.Errorf("export should include the chef storefront: %+v", export.Chef)
	}
	if export.Orders == nil || export.Reviews == nil || export.Addresses == nil || export.Conversations == nil {
		t.Errorf("export collections must be non-nil: %+v", export)
	}
}

func TestAccountService_Export_NoChefIsFine(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	u := &domain.User{Email: "cust@b.c", PasswordHash: "x", IsActive: true}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	svc := newAccountService(repo, newFakeChefRepo(), &fakeAccountRepo{})

	export, err := svc.Export(ctx, u.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if export.Chef != nil {
		t.Errorf("customer export should have no chef: %+v", export.Chef)
	}
}
