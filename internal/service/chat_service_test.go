package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeChatRepo is an in-memory domain.ChatRepository for tests.
type fakeChatRepo struct {
	mu       sync.Mutex
	convs    map[int]*domain.Conversation
	msgs     []*domain.Message
	nextConv int
	nextMsg  int
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{convs: map[int]*domain.Conversation{}, nextConv: 1, nextMsg: 1}
}

func (f *fakeChatRepo) FindConversation(_ context.Context, userID, chefID int) (*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.convs {
		if c.UserID == userID && c.ChefID == chefID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrConversationNotFound
}
func (f *fakeChatRepo) FindConversationByID(_ context.Context, id int) (*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if c, ok := f.convs[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, domain.ErrConversationNotFound
}
func (f *fakeChatRepo) CreateConversation(_ context.Context, c *domain.Conversation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c.ID = f.nextConv
	f.nextConv++
	c.CreatedAt = time.Now()
	cp := *c
	f.convs[c.ID] = &cp
	return nil
}
func (f *fakeChatRepo) ListConversationsByUser(_ context.Context, userID int) ([]*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Conversation, 0)
	for _, c := range f.convs {
		if c.UserID == userID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeChatRepo) ListConversationsByChef(_ context.Context, chefID int) ([]*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Conversation, 0)
	for _, c := range f.convs {
		if c.ChefID == chefID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeChatRepo) CreateMessage(_ context.Context, m *domain.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m.ID = f.nextMsg
	f.nextMsg++
	m.CreatedAt = time.Now()
	cp := *m
	f.msgs = append(f.msgs, &cp)
	if c, ok := f.convs[m.ConversationID]; ok {
		c.LastMessageAt = &m.CreatedAt
	}
	return nil
}
func (f *fakeChatRepo) ListMessages(_ context.Context, conversationID, limit, offset int) ([]*domain.Message, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Message, 0)
	for _, m := range f.msgs {
		if m.ConversationID == conversationID {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}

// chatFixture wires a ChatService over fakes, seeding chef profiles for the
// given user ids (chef.ID assigned in order).
func chatFixture(t *testing.T, chefUserIDs ...int) (*service.ChatService, *fakeChatRepo) {
	t.Helper()
	chefRepo := newFakeChefRepo()
	for _, uid := range chefUserIDs {
		if err := chefRepo.Create(context.Background(), &domain.Chef{UserID: uid, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}
	chats := newFakeChatRepo()
	return service.NewChatService(chats, chefRepo), chats
}

func TestChatService_StartIsIdempotent(t *testing.T) {
	svc, _ := chatFixture(t, 1) // user1 -> chef1
	ctx := context.Background()

	first, err := svc.StartConversation(ctx, 100, 1, nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	again, err := svc.StartConversation(ctx, 100, 1, nil)
	if err != nil {
		t.Fatalf("start again: %v", err)
	}
	if first.ID != again.ID {
		t.Errorf("expected the same conversation, got %d and %d", first.ID, again.ID)
	}
}

func TestChatService_StartUnknownChef(t *testing.T) {
	svc, _ := chatFixture(t) // no chefs
	if _, err := svc.StartConversation(context.Background(), 100, 99, nil); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("err = %v, want ErrChefNotFound", err)
	}
}

func TestChatService_SendAndHistoryAuthorization(t *testing.T) {
	svc, _ := chatFixture(t, 1) // user1 -> chef1
	ctx := context.Background()
	conv, err := svc.StartConversation(ctx, 100, 1, nil) // customer 100 with chef1
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Customer can send.
	if _, err := svc.SendMessage(ctx, 100, conv.ID, "hello chef"); err != nil {
		t.Fatalf("customer send: %v", err)
	}
	// The chef (user 1, owning chef 1) can also send.
	if _, err := svc.SendMessage(ctx, 1, conv.ID, "hi customer"); err != nil {
		t.Fatalf("chef send: %v", err)
	}
	// A stranger cannot send or read.
	if _, err := svc.SendMessage(ctx, 999, conv.ID, "intrude"); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("stranger send = %v, want ErrForbidden", err)
	}
	if _, _, err := svc.Messages(ctx, 999, conv.ID, 50, 0); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("stranger read = %v, want ErrForbidden", err)
	}

	// History loads for a participant, in order.
	msgs, total, err := svc.Messages(ctx, 100, conv.ID, 50, 0)
	if total != 2 {
		t.Errorf("messages total = %d, want 2", total)
	}
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(msgs) != 2 || msgs[0].Body != "hello chef" {
		t.Errorf("unexpected history: %+v", msgs)
	}
}

func TestChatService_EmptyMessageRejected(t *testing.T) {
	svc, _ := chatFixture(t, 1)
	ctx := context.Background()
	conv, _ := svc.StartConversation(ctx, 100, 1, nil)
	if _, err := svc.SendMessage(ctx, 100, conv.ID, "   "); !errors.Is(err, domain.ErrEmptyMessage) {
		t.Errorf("empty message = %v, want ErrEmptyMessage", err)
	}
}

func (f *fakeChatRepo) MarkRead(_ context.Context, conversationID, readerUserID int) error {
	now := time.Now()
	for _, m := range f.msgs {
		if m.ConversationID == conversationID && m.SenderID != readerUserID && m.ReadAt == nil {
			m.ReadAt = &now
		}
	}
	return nil
}
