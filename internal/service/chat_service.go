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

// SendMessage posts a message to a conversation the sender participates in.
func (s *ChatService) SendMessage(ctx context.Context, senderUserID, conversationID int, body string) (*domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, domain.ErrEmptyMessage
	}
	if _, err := s.authorize(ctx, senderUserID, conversationID); err != nil {
		return nil, err
	}

	msg := &domain.Message{ConversationID: conversationID, SenderID: senderUserID, Body: body}
	if err := s.chats.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Messages returns a conversation's history for a participant.
func (s *ChatService) Messages(ctx context.Context, requesterUserID, conversationID, limit, offset int) ([]*domain.Message, error) {
	if _, err := s.authorize(ctx, requesterUserID, conversationID); err != nil {
		return nil, err
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
func (s *ChatService) Authorize(ctx context.Context, requesterUserID, conversationID int) (*domain.Conversation, error) {
	return s.authorize(ctx, requesterUserID, conversationID)
}

func (s *ChatService) authorize(ctx context.Context, requesterUserID, conversationID int) (*domain.Conversation, error) {
	conv, err := s.chats.FindConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	chefID, err := s.chefIDFor(ctx, requesterUserID)
	if err != nil {
		return nil, err
	}
	if !conv.IsParticipant(requesterUserID, chefID) {
		return nil, domain.ErrForbidden
	}
	return conv, nil
}
