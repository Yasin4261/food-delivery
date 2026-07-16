package domain

import "time"

// AccountExport is the machine-readable dump of everything the platform holds
// about one account (#107, GDPR data portability). It is assembled by the
// service from the caller's own records — orders they placed, reviews they
// wrote, chat threads they took part in — never another user's data.
type AccountExport struct {
	ExportedAt    time.Time             `json:"exported_at"`
	User          *User                 `json:"user"`
	Chef          *Chef                 `json:"chef,omitempty"`
	Addresses     []*Address            `json:"addresses"`
	Orders        []*Order              `json:"orders"`
	Reviews       []*Review             `json:"reviews"`
	Conversations []*ConversationExport `json:"conversations"`
}

// ConversationExport is one chat thread the account took part in, with its
// messages, for the data export.
type ConversationExport struct {
	Conversation *Conversation `json:"conversation"`
	Messages     []*Message    `json:"messages"`
}
