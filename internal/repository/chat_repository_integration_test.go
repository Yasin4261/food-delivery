//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestChatRepository_ConversationAndMessages(t *testing.T) {
	resetDB(t)
	repo := repository.NewChatRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)

	conv := &domain.Conversation{UserID: customer.ID, ChefID: chef.ID}
	if err := repo.CreateConversation(ctx(), conv); err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if conv.ID == 0 {
		t.Fatal("create did not back-fill id")
	}

	// Lookups.
	if got, err := repo.FindConversation(ctx(), customer.ID, chef.ID); err != nil || got.ID != conv.ID {
		t.Errorf("find conversation = %+v, %v", got, err)
	}
	if byUser, _ := repo.ListConversationsByUser(ctx(), customer.ID); len(byUser) != 1 {
		t.Errorf("list by user = %d, want 1", len(byUser))
	}
	if byChef, _ := repo.ListConversationsByChef(ctx(), chef.ID); len(byChef) != 1 {
		t.Errorf("list by chef = %d, want 1", len(byChef))
	}

	// Messages persist and load as history (oldest first).
	if err := repo.CreateMessage(ctx(), &domain.Message{ConversationID: conv.ID, SenderID: customer.ID, Body: "first"}); err != nil {
		t.Fatalf("create message 1: %v", err)
	}
	if err := repo.CreateMessage(ctx(), &domain.Message{ConversationID: conv.ID, SenderID: chef.UserID, Body: "second"}); err != nil {
		t.Fatalf("create message 2: %v", err)
	}

	msgs, _, err := repo.ListMessages(ctx(), conv.ID, 50, 0)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 2 || msgs[0].Body != "first" || msgs[1].Body != "second" {
		t.Errorf("unexpected history: %+v", msgs)
	}

	// last_message_at was bumped on the conversation.
	got, _ := repo.FindConversationByID(ctx(), conv.ID)
	if got.LastMessageAt == nil {
		t.Error("last_message_at should be set after messages")
	}

	if _, err := repo.FindConversationByID(ctx(), 999); !errors.Is(err, domain.ErrConversationNotFound) {
		t.Errorf("missing conversation = %v, want ErrConversationNotFound", err)
	}
}

func TestChatRepository_UnreadAndMarkRead(t *testing.T) {
	resetDB(t)
	repo := repository.NewChatRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chefUser := seedUser(t, "chef@example.com")
	chef := seedChef(t, chefUser.ID)

	conv := &domain.Conversation{UserID: customer.ID, ChefID: chef.ID}
	if err := repo.CreateConversation(ctx(), conv); err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	// Customer sends 2, chef sends 1.
	for _, m := range []*domain.Message{
		{ConversationID: conv.ID, SenderID: customer.ID, Body: "c1"},
		{ConversationID: conv.ID, SenderID: customer.ID, Body: "c2"},
		{ConversationID: conv.ID, SenderID: chefUser.ID, Body: "reply"},
	} {
		if err := repo.CreateMessage(ctx(), m); err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	// Chef's unread = the 2 customer messages; customer's unread = the 1 reply.
	chefList, _ := repo.ListConversationsByChef(ctx(), chef.ID)
	if len(chefList) != 1 || chefList[0].UnreadCount != 2 {
		t.Errorf("chef unread = %d, want 2", chefList[0].UnreadCount)
	}
	custList, _ := repo.ListConversationsByUser(ctx(), customer.ID)
	if custList[0].UnreadCount != 1 {
		t.Errorf("customer unread = %d, want 1", custList[0].UnreadCount)
	}

	// Chef marks read -> the customer's messages are read; the chef's unread
	// drops to 0, but the customer's unread (the chef's reply) is unchanged.
	if err := repo.MarkRead(ctx(), conv.ID, chefUser.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	chefList, _ = repo.ListConversationsByChef(ctx(), chef.ID)
	if chefList[0].UnreadCount != 0 {
		t.Errorf("chef unread after read = %d, want 0", chefList[0].UnreadCount)
	}
	custList, _ = repo.ListConversationsByUser(ctx(), customer.ID)
	if custList[0].UnreadCount != 1 {
		t.Errorf("customer unread after chef read = %d, want 1 (unaffected)", custList[0].UnreadCount)
	}
}
