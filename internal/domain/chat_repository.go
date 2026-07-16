package domain

import "context"

// ChatRepository is the port for chat persistence. Lookups return
// ErrConversationNotFound when no row matches.
type ChatRepository interface {
	// FindConversation returns the thread for a (userID, chefID) pair, or
	// ErrConversationNotFound.
	FindConversation(ctx context.Context, userID, chefID int) (*Conversation, error)
	FindConversationByID(ctx context.Context, id int) (*Conversation, error)
	CreateConversation(ctx context.Context, c *Conversation) error
	// ListConversationsByUser / ByChef return a participant's threads, most
	// recently active first.
	ListConversationsByUser(ctx context.Context, userID int) ([]*Conversation, error)
	ListConversationsByChef(ctx context.Context, chefID int) ([]*Conversation, error)

	// CreateMessage persists a message and bumps the conversation's
	// last_message_at in the same transaction.
	CreateMessage(ctx context.Context, m *Message) error
	// ListMessages returns a page of a conversation's messages (oldest first)
	// and the total message count.
	ListMessages(ctx context.Context, conversationID, limit, offset int) ([]*Message, int, error)
	// MarkRead stamps read_at on the conversation's messages NOT sent by
	// readerUserID that are still unread.
	MarkRead(ctx context.Context, conversationID, readerUserID int) error
}
