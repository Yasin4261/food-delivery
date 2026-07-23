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

// Support threads (#120) against real Postgres: the migration relaxes chef_id
// and the partial unique indexes must hold — one support thread per user, chef
// threads unaffected — and the kind-shape CHECK must reject malformed rows.
func TestChatRepository_SupportConversations(t *testing.T) {
	resetDB(t)
	repo := repository.NewChatRepository(testDB)
	user := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)

	// A support thread stores chef_id NULL and reads back as ChefID 0.
	sup := &domain.Conversation{Kind: domain.ConversationKindSupport, UserID: user.ID}
	if err := repo.CreateConversation(ctx(), sup); err != nil {
		t.Fatalf("create support: %v", err)
	}
	got, err := repo.FindSupportConversation(ctx(), user.ID)
	if err != nil {
		t.Fatalf("find support: %v", err)
	}
	if !got.IsSupport() || got.ChefID != 0 || got.ID != sup.ID {
		t.Fatalf("support round-trip wrong: %+v", got)
	}

	// One support thread per user (partial unique index).
	dup := &domain.Conversation{Kind: domain.ConversationKindSupport, UserID: user.ID}
	if err := repo.CreateConversation(ctx(), dup); err == nil {
		t.Error("a second support thread for the same user should violate the unique index")
	}

	// A chef thread for the same user coexists (different partial index).
	chefConv := &domain.Conversation{Kind: domain.ConversationKindChef, UserID: user.ID, ChefID: chef.ID}
	if err := repo.CreateConversation(ctx(), chefConv); err != nil {
		t.Fatalf("create chef thread alongside support: %v", err)
	}

	// FindConversation returns only the chef thread; the support thread is not
	// a (user, chef) match.
	if c, err := repo.FindConversation(ctx(), user.ID, chef.ID); err != nil || c.ID != chefConv.ID {
		t.Errorf("find chef thread = %+v, %v", c, err)
	}

	// The support inbox lists the support thread with an admin-side unread
	// count (messages from the target user).
	if err := repo.CreateMessage(ctx(), &domain.Message{ConversationID: sup.ID, SenderID: user.ID, Body: "help"}); err != nil {
		t.Fatalf("seed support message: %v", err)
	}
	inbox, err := repo.ListSupportConversations(ctx())
	if err != nil || len(inbox) != 1 || inbox[0].ID != sup.ID {
		t.Fatalf("inbox = %+v, %v, want the one support thread", inbox, err)
	}
	if inbox[0].UnreadCount != 1 {
		t.Errorf("admin-side unread = %d, want 1 (the user's message)", inbox[0].UnreadCount)
	}

	// The user sees the support thread in their own list too (customer-side
	// unread = messages from the admin; none yet).
	byUser, err := repo.ListConversationsByUser(ctx(), user.ID)
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	var sawSupport bool
	for _, c := range byUser {
		if c.ID == sup.ID {
			sawSupport = true
			if c.UnreadCount != 0 {
				t.Errorf("user-side unread = %d, want 0", c.UnreadCount)
			}
		}
	}
	if !sawSupport {
		t.Error("the user should see their support thread in ListConversationsByUser")
	}

	// The kind-shape CHECK rejects a support row that names a kitchen, and a
	// chef row with no kitchen.
	if _, err := testDB.Exec(
		`INSERT INTO chat_conversations (kind, user_id, chef_id) VALUES ('support', $1, $2)`, user.ID, chef.ID); err == nil {
		t.Error("a support thread with a chef_id should violate the kind-shape check")
	}
	if _, err := testDB.Exec(
		`INSERT INTO chat_conversations (kind, user_id, chef_id) VALUES ('chef', $1, NULL)`, user.ID); err == nil {
		t.Error("a chef thread without a chef_id should violate the kind-shape check")
	}
}
