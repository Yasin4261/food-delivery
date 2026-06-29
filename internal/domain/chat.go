package domain

import "time"

// Conversation is a message thread between a customer (UserID) and a chef
// (ChefID), optionally tied to an order (mirrors chat_conversations).
type Conversation struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	ChefID        int        `json:"chef_id"`
	OrderID       *int       `json:"order_id,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// IsParticipant reports whether the requester may access the conversation. The
// requester is identified by their user id and, if they are a chef, their chef
// profile id (pass 0 when the requester has no chef profile). Only the customer
// (UserID) and the owning chef (ChefID) are participants.
func (c *Conversation) IsParticipant(userID, chefID int) bool {
	if userID == c.UserID {
		return true
	}
	return chefID != 0 && chefID == c.ChefID
}

// Message is a single message in a conversation (mirrors chat_messages).
type Message struct {
	ID             int        `json:"id"`
	ConversationID int        `json:"conversation_id"`
	SenderID       int        `json:"sender_id"`
	Body           string     `json:"body"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
