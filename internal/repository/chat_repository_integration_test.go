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
