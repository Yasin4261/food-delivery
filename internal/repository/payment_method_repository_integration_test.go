//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestPaymentMethodRepository_RoundTrip(t *testing.T) {
	resetDB(t)
	repo := repository.NewPaymentMethodRepository(testDB)
	user := seedUser(t, "cards@example.com")
	other := seedUser(t, "other@example.com")

	// No cards, no wallet key yet.
	if list, err := repo.ListByUser(ctx(), user.ID); err != nil || len(list) != 0 {
		t.Fatalf("initial list = %+v, %v, want empty", list, err)
	}
	if key, err := repo.CardUserKey(ctx(), user.ID); err != nil || key != "" {
		t.Fatalf("initial card user key = %q, %v, want empty", key, err)
	}

	card := &domain.SavedCard{
		UserID:       user.ID,
		CardUserKey:  "cuk-1",
		CardToken:    "ctok-1",
		MaskedNumber: "552608******0006",
		Association:  "MASTER_CARD",
		Family:       "Bonus",
		BankName:     "Test Bank",
	}
	if err := repo.Add(ctx(), card); err != nil {
		t.Fatalf("add: %v", err)
	}
	if card.ID == 0 || card.CreatedAt.IsZero() {
		t.Fatalf("add did not back-fill id/created_at: %+v", card)
	}
	firstID := card.ID

	// Add is idempotent on (user, token).
	dup := &domain.SavedCard{UserID: user.ID, CardUserKey: "cuk-1", CardToken: "ctok-1", MaskedNumber: "552608******0006"}
	if err := repo.Add(ctx(), dup); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	list, err := repo.ListByUser(ctx(), user.ID)
	if err != nil || len(list) != 1 || list[0].ID != firstID {
		t.Fatalf("after dup add list = %+v, %v, want 1 unchanged row", list, err)
	}
	if list[0].Association != "MASTER_CARD" || list[0].Family != "Bonus" || list[0].BankName != "Test Bank" {
		t.Errorf("metadata round-trip wrong: %+v", list[0])
	}

	// A second card, newest first.
	if err := repo.Add(ctx(), &domain.SavedCard{UserID: user.ID, CardUserKey: "cuk-1", CardToken: "ctok-2", MaskedNumber: "411111******1111"}); err != nil {
		t.Fatalf("add 2: %v", err)
	}
	list, _ = repo.ListByUser(ctx(), user.ID)
	if len(list) != 2 || list[0].CardToken != "ctok-2" {
		t.Fatalf("list = %+v, want 2 newest-first", list)
	}

	// Wallet key resolves; find by token is owner-scoped.
	if key, _ := repo.CardUserKey(ctx(), user.ID); key != "cuk-1" {
		t.Errorf("card user key = %q, want cuk-1", key)
	}
	if _, err := repo.FindByToken(ctx(), other.ID, "ctok-1"); !errors.Is(err, domain.ErrCardNotFound) {
		t.Errorf("foreign find = %v, want ErrCardNotFound", err)
	}

	// Delete is owner-scoped.
	if err := repo.Delete(ctx(), other.ID, "ctok-1"); !errors.Is(err, domain.ErrCardNotFound) {
		t.Errorf("foreign delete = %v, want ErrCardNotFound", err)
	}
	if err := repo.Delete(ctx(), user.ID, "ctok-1"); err != nil {
		t.Fatalf("delete own: %v", err)
	}
	list, _ = repo.ListByUser(ctx(), user.ID)
	if len(list) != 1 || list[0].CardToken != "ctok-2" {
		t.Errorf("after delete list = %+v, want only ctok-2", list)
	}
}
