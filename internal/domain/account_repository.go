package domain

import "context"

// AccountRepository is the port for account-lifecycle persistence that spans
// several tables in one transaction (#107). The Postgres adapter lives in
// internal/repository.
type AccountRepository interface {
	// Anonymise irreversibly scrubs the user's personal data and deactivates
	// the account (and their chef storefront, if any) in a single transaction:
	// PII on users/chefs is cleared, addresses are deleted and chat message
	// bodies are replaced with a tombstone. Counterparty records — orders,
	// order_items, sub_orders and reviews — are retained so the other party's
	// history and earnings stay intact. Login is blocked afterwards
	// (is_active = false).
	Anonymise(ctx context.Context, userID int) error
}
