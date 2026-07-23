package domain

import (
	"encoding/json"
	"time"
)

// Audit action names — one per admin mutation. The before/after JSON carries
// the concrete values that changed.
const (
	AuditUserSetActive    = "user.set_active"
	AuditChefSetActive    = "chef.set_active"
	AuditChefSetOnline    = "chef.set_online"
	AuditChefSetAccepting = "chef.set_accepting"
	AuditPromoCreate      = "promo.create"
	AuditPromoUpdate      = "promo.update"
	AuditPromoDelete      = "promo.delete"
	AuditPromoSetActive   = "promo.set_active"
)

// Audit target types.
const (
	AuditTargetUser  = "user"
	AuditTargetChef  = "chef"
	AuditTargetOrder = "order"
	AuditTargetPromo = "promo"
)

// AuditEntry is one row of the admin audit log (mirrors admin_audit_log,
// migration 000027). It is written in the same transaction as the mutation it
// records. Before/After hold the prior/new state as JSON (nil for
// creates/deletes respectively); they must never contain secrets (no password
// hashes, tokens or card data — CLAUDE.md §10 A09).
type AuditEntry struct {
	ID          int             `json:"id"`
	ActorUserID int             `json:"actor_user_id"`
	Action      string          `json:"action"`
	TargetType  string          `json:"target_type"`
	TargetID    int             `json:"target_id"`
	Reason      string          `json:"reason,omitempty"`
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// AuditFilters narrows the audit log listing. The zero value matches every
// entry.
type AuditFilters struct {
	Action     string
	TargetType string
	TargetID   int // 0 = any
	ActorID    int // 0 = any
}
