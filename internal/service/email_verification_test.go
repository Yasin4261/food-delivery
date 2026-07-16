package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeVerificationRepo is an in-memory domain.EmailVerificationRepository.
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

// newServiceWithVerification wires the verification flow onto a fresh service.
func newServiceWithVerification(repo domain.UserRepository, verifs domain.EmailVerificationRepository, mail domain.Mailer) *service.AuthService {
	s := service.NewAuthService(repo, newFakeResetRepo(), mail, "test-secret", time.Hour, "http://app.test")
	s.SetEmailVerification(verifs)
	return s
}

func TestRegister_SendsVerificationEmail(t *testing.T) {
	repo := newFakeUserRepo()
	mail := &recordingMailer{}
	svc := newServiceWithVerification(repo, newFakeVerificationRepo(), mail)

	res, err := svc.Register(context.Background(), service.RegisterInput{
		Username: "alice", Email: "alice@example.com", Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.User.IsVerified {
		t.Fatal("newly registered user must not be verified")
	}
	msg, ok := mail.last()
	if !ok {
		t.Fatal("expected a verification email to be sent")
	}
	if msg.To != "alice@example.com" {
		t.Fatalf("email sent to %q, want alice@example.com", msg.To)
	}
	if extractToken(t, msg.Body) == "" {
		t.Fatal("verification email carried no token")
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	repo := newFakeUserRepo()
	mail := &recordingMailer{}
	svc := newServiceWithVerification(repo, newFakeVerificationRepo(), mail)

	if _, err := svc.Register(context.Background(), service.RegisterInput{
		Username: "bob", Email: "bob@example.com", Password: "hunter2",
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	msg, _ := mail.last()
	token := extractToken(t, msg.Body)

	if err := svc.VerifyEmail(context.Background(), token); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	user, _ := repo.FindByEmail(context.Background(), "bob@example.com")
	if !user.IsVerified {
		t.Fatal("account should be verified after redeeming the token")
	}

	// The token is single-use: replaying it fails.
	if err := svc.VerifyEmail(context.Background(), token); err != domain.ErrInvalidVerificationToken {
		t.Fatalf("replayed token: got %v, want ErrInvalidVerificationToken", err)
	}
}

func TestVerifyEmail_BadToken(t *testing.T) {
	svc := newServiceWithVerification(newFakeUserRepo(), newFakeVerificationRepo(), &recordingMailer{})
	if err := svc.VerifyEmail(context.Background(), "not-a-real-token"); err != domain.ErrInvalidVerificationToken {
		t.Fatalf("got %v, want ErrInvalidVerificationToken", err)
	}
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	repo := newFakeUserRepo()
	mail := &recordingMailer{}
	svc := newServiceWithVerification(repo, newFakeVerificationRepo(), mail)

	res, err := svc.Register(context.Background(), service.RegisterInput{
		Username: "carol", Email: "carol@example.com", Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := repo.MarkVerified(context.Background(), res.User.ID); err != nil {
		t.Fatalf("MarkVerified: %v", err)
	}

	if err := svc.ResendVerification(context.Background(), res.User.ID); err != domain.ErrAlreadyVerified {
		t.Fatalf("got %v, want ErrAlreadyVerified", err)
	}
}

func TestResendVerification_SendsFreshLink(t *testing.T) {
	repo := newFakeUserRepo()
	mail := &recordingMailer{}
	svc := newServiceWithVerification(repo, newFakeVerificationRepo(), mail)

	res, err := svc.Register(context.Background(), service.RegisterInput{
		Username: "dave", Email: "dave@example.com", Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	before := len(mail.sent)

	if err := svc.ResendVerification(context.Background(), res.User.ID); err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}
	if len(mail.sent) != before+1 {
		t.Fatalf("expected one more email, sent went %d -> %d", before, len(mail.sent))
	}

	// The freshly-issued token verifies the account.
	msg, _ := mail.last()
	if err := svc.VerifyEmail(context.Background(), extractToken(t, msg.Body)); err != nil {
		t.Fatalf("VerifyEmail with resent token: %v", err)
	}
}
