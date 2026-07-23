package service

import (
	"context"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChatService implements messaging between a customer and a chef. Only the two
// participants of a conversation may read it or post to it.
type ChatService struct {
	chats domain.ChatRepository
	chefs domain.ChefRepository
}

// NewChatService builds a ChatService.
func NewChatService(chats domain.ChatRepository, chefs domain.ChefRepository) *ChatService {
	return &ChatService{chats: chats, chefs: chefs}
}

// chefIDFor returns the requester's chef profile id, or 0 if they have none.
func (s *ChatService) chefIDFor(ctx context.Context, userID int) (int, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err == domain.ErrChefNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return chef.ID, nil
}

// StartConversation opens (or returns the existing) thread between the calling
// customer and a chef. The chef must exist.
func (s *ChatService) StartConversation(ctx context.Context, customerUserID, chefID int, orderID *int) (*domain.Conversation, error) {
	if _, err := s.chefs.FindByID(ctx, chefID); err != nil {
		return nil, err
	}

	existing, err := s.chats.FindConversation(ctx, customerUserID, chefID)
	if err == nil {
		return existing, nil
	} else if err != domain.ErrConversationNotFound {
		return nil, err
	}

	conv := &domain.Conversation{UserID: customerUserID, ChefID: chefID, OrderID: orderID}
	if err := s.chats.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// StartSupportConversation opens (or returns the existing) support thread for a
// target user. Idempotent: one support thread per user. Called by an admin
// (opening a thread with someone) or by a user contacting support about
// themselves — in both cases targetUserID is the non-admin party.
func (s *ChatService) StartSupportConversation(ctx context.Context, targetUserID int) (*domain.Conversation, error) {
	existing, err := s.chats.FindSupportConversation(ctx, targetUserID)
	if err == nil {
		return existing, nil
	} else if err != domain.ErrConversationNotFound {
		return nil, err
	}
	conv := &domain.Conversation{Kind: domain.ConversationKindSupport, UserID: targetUserID}
	if err := s.chats.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// SupportConversations lists every support thread — the admin inbox.
func (s *ChatService) SupportConversations(ctx context.Context) ([]*domain.Conversation, error) {
	return s.chats.ListSupportConversations(ctx)
}

// SendMessage posts a message to a conversation the sender participates in.
// isAdmin marks the sender as an admin, which makes them a participant of a
// support thread (but never of a customer<->chef thread).
func (s *ChatService) SendMessage(ctx context.Context, senderUserID int, isAdmin bool, conversationID int, body string) (*domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, domain.ErrEmptyMessage
	}
	if _, err := s.authorize(ctx, senderUserID, isAdmin, conversationID); err != nil {
		return nil, err
	}

	msg := &domain.Message{ConversationID: conversationID, SenderID: senderUserID, Body: body}
	if err := s.chats.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Messages returns a page of a conversation's history for a participant and the
// total message count.
func (s *ChatService) Messages(ctx context.Context, requesterUserID int, isAdmin bool, conversationID, limit, offset int) ([]*domain.Message, int, error) {
	if _, err := s.authorize(ctx, requesterUserID, isAdmin, conversationID); err != nil {
		return nil, 0, err
	}
	limit, offset = normalisePaging(limit, offset)
	return s.chats.ListMessages(ctx, conversationID, limit, offset)
}

// Conversations lists the threads the requester participates in (as customer
// and, if they have a chef profile, as chef).
func (s *ChatService) Conversations(ctx context.Context, requesterUserID int) ([]*domain.Conversation, error) {
	out, err := s.chats.ListConversationsByUser(ctx, requesterUserID)
	if err != nil {
		return nil, err
	}
	chefID, err := s.chefIDFor(ctx, requesterUserID)
	if err != nil {
		return nil, err
	}
	if chefID != 0 {
		asChef, err := s.chats.ListConversationsByChef(ctx, chefID)
		if err != nil {
			return nil, err
		}
		out = append(out, asChef...)
	}
	return out, nil
}

// Authorize returns the conversation if the requester is a participant,
// otherwise ErrForbidden (or ErrConversationNotFound). It is exported via the
// thin wrapper used by the WebSocket transport.
func (s *ChatService) Authorize(ctx context.Context, requesterUserID int, isAdmin bool, conversationID int) (*domain.Conversation, error) {
	return s.authorize(ctx, requesterUserID, isAdmin, conversationID)
}

// MarkRead marks the other party's messages in a conversation as read for the
// requester (participant-only).
func (s *ChatService) MarkRead(ctx context.Context, requesterUserID int, isAdmin bool, conversationID int) error {
	if _, err := s.authorize(ctx, requesterUserID, isAdmin, conversationID); err != nil {
		return err
	}
	return s.chats.MarkRead(ctx, conversationID, requesterUserID)
}

func (s *ChatService) authorize(ctx context.Context, requesterUserID int, isAdmin bool, conversationID int) (*domain.Conversation, error) {
	conv, err := s.chats.FindConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	// A chef profile only matters for chef threads; skip the lookup for
	// support threads (and for admins, who are participants by role).
	chefID := 0
	if !conv.IsSupport() {
		chefID, err = s.chefIDFor(ctx, requesterUserID)
		if err != nil {
			return nil, err
		}
	}
	if !conv.IsParticipant(requesterUserID, chefID, isAdmin) {
		return nil, domain.ErrForbidden
	}
	return conv, nil
}
