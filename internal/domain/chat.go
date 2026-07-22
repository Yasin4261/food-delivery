package domain

import "time"

// Conversation kinds. A chef thread is the original customer<->chef shape; a
// support thread is admin<->user (#120).
const (
	ConversationKindChef    = "chef"
	ConversationKindSupport = "support"
)

// Conversation is a message thread. For a chef thread (Kind == "chef") it is
// between a customer (UserID) and a chef (ChefID). For a support thread
// (Kind == "support") it is between a user (UserID — the customer or chef being
// helped) and any admin, and ChefID is 0. Mirrors chat_conversations.
type Conversation struct {
	ID            int        `json:"id"`
	Kind          string     `json:"kind"`
	UserID        int        `json:"user_id"`
	ChefID        int        `json:"chef_id,omitempty"` // 0 on a support thread
	OrderID       *int       `json:"order_id,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	// UnreadCount is the requester's unread messages in this thread (messages
	// from the other party with read_at NULL). Derived at list time, not a
	// column.
	UnreadCount int `json:"unread_count"`
}

// IsSupport reports whether this is an admin<->user support thread.
func (c *Conversation) IsSupport() bool { return c.Kind == ConversationKindSupport }

// IsParticipant reports whether the requester may access the conversation. The
// requester is identified by their user id, their chef profile id (0 when they
// have none), and whether they are an admin.
//
//   - Support thread: the target user (UserID) or ANY admin. A chef profile is
//     irrelevant here.
//   - Chef thread: the customer (UserID) or the owning chef (ChefID). Admins are
//     deliberately NOT participants — support must never read a customer<->chef
//     thread through this door (#120).
func (c *Conversation) IsParticipant(userID, chefID int, isAdmin bool) bool {
	if c.IsSupport() {
		return userID == c.UserID || isAdmin
	}
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
