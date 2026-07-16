//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestAccountRepository_Anonymise(t *testing.T) {
	resetDB(t)
	accountRepo := repository.NewAccountRepository(testDB)
	userRepo := repository.NewUserRepository(testDB)
	chefRepo := repository.NewChefRepository(testDB)
	addressRepo := repository.NewAddressRepository(testDB)
	orderRepo := repository.NewOrderRepository(testDB)
	chatRepo := repository.NewChatRepository(testDB)

	// A chef account with a storefront, an address, an order and a chat message.
	user := seedUser(t, "gdpr@example.com")
	phone := "555-1234"
	user.PhoneNumber = &phone
	if err := userRepo.UpdateProfile(ctx(), user); err != nil {
		t.Fatalf("set phone: %v", err)
	}
	chef := seedChef(t, user.ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 10, 100)

	addr := &domain.Address{UserID: user.ID, Label: "Home", Address: "1 Privacy Ln", IsDefault: true}
	if err := addressRepo.Create(ctx(), addr); err != nil {
		t.Fatalf("seed address: %v", err)
	}

	// A second customer places an order with this chef — counterparty history
	// that must survive the chef's deletion.
	buyer := seedUser(t, "buyer@example.com")
	order := buildOrder(buyer.ID, "ORD-GDPR",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, item.Price))
	if err := orderRepo.Create(ctx(), order); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	conv := &domain.Conversation{UserID: buyer.ID, ChefID: chef.ID}
	if err := chatRepo.CreateConversation(ctx(), conv); err != nil {
		t.Fatalf("seed conversation: %v", err)
	}
	msg := &domain.Message{ConversationID: conv.ID, SenderID: user.ID, Body: "secret note"}
	if err := chatRepo.CreateMessage(ctx(), msg); err != nil {
		t.Fatalf("seed message: %v", err)
	}

	// Anonymise the chef's account.
	if err := accountRepo.Anonymise(ctx(), user.ID); err != nil {
		t.Fatalf("anonymise: %v", err)
	}

	// User PII is scrubbed and the account deactivated.
	got, err := userRepo.FindByID(ctx(), user.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if got.IsActive {
		t.Error("account should be deactivated")
	}
	if got.Email == "gdpr@example.com" || got.PhoneNumber != nil || got.PasswordHash != "" {
		t.Errorf("PII not scrubbed: %+v", got)
	}

	// The old email is now free for a fresh sign-up.
	if _, err := userRepo.FindByEmail(ctx(), "gdpr@example.com"); err != domain.ErrUserNotFound {
		t.Errorf("old email should no longer resolve: %v", err)
	}

	// Chef storefront scrubbed + hidden.
	gotChef, err := chefRepo.FindByUserID(ctx(), user.ID)
	if err != nil {
		t.Fatalf("find chef: %v", err)
	}
	if gotChef.BusinessName != "Closed kitchen" || gotChef.IsActive {
		t.Errorf("chef not scrubbed/deactivated: %+v", gotChef)
	}

	// Personal-only rows are gone.
	if addrs, _ := addressRepo.ListByUser(ctx(), user.ID); len(addrs) != 0 {
		t.Errorf("addresses should be deleted: %+v", addrs)
	}

	// Counterparty order survives (the buyer's history / chef's earnings basis).
	if _, err := orderRepo.FindByID(ctx(), order.ID); err != nil {
		t.Errorf("counterparty order must be retained: %v", err)
	}

	// Chat thread retained, but the message body is tombstoned.
	msgs, _, err := chatRepo.ListMessages(ctx(), conv.ID, 10, 0)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Body != "[deleted]" {
		t.Errorf("message body not tombstoned: %+v", msgs)
	}
}

func TestReviewRepository_ListByUser(t *testing.T) {
	resetDB(t)
	reviewRepo := repository.NewReviewRepository(testDB)
	orderRepo := repository.NewOrderRepository(testDB)

	buyer := seedUser(t, "rev@example.com")
	chef := seedChef(t, seedUser(t, "chef2@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 10, 100)

	order := buildOrder(buyer.ID, "ORD-REV",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, item.Price))
	if err := orderRepo.Create(ctx(), order); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	comment := "great"
	rv := &domain.Review{UserID: buyer.ID, OrderID: order.ID, ChefID: &chef.ID, Rating: 5, Comment: &comment}
	if err := reviewRepo.Create(ctx(), rv); err != nil {
		t.Fatalf("seed review: %v", err)
	}

	got, err := reviewRepo.ListByUser(ctx(), buyer.ID)
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if len(got) != 1 || got[0].Rating != 5 {
		t.Errorf("ListByUser = %+v, want one 5-star review", got)
	}
	// A different user sees none of them.
	if other, _ := reviewRepo.ListByUser(ctx(), buyer.ID+999); len(other) != 0 {
		t.Errorf("foreign user should have no reviews: %+v", other)
	}
}
