package domain_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func TestConversation_IsParticipant(t *testing.T) {
	conv := &domain.Conversation{Kind: domain.ConversationKindChef, UserID: 10, ChefID: 3}

	// The customer (by user id).
	if !conv.IsParticipant(10, 0, false) {
		t.Error("customer should be a participant")
	}
	// The owning chef (by chef id), even though their user id differs.
	if !conv.IsParticipant(99, 3, false) {
		t.Error("owning chef should be a participant")
	}
	// An unrelated user with no matching chef id.
	if conv.IsParticipant(42, 0, false) {
		t.Error("unrelated user should not be a participant")
	}
	// A different chef.
	if conv.IsParticipant(42, 7, false) {
		t.Error("a different chef should not be a participant")
	}
	// An admin is NOT a participant of a chef thread — support must never read
	// customer<->chef conversations through the admin door (#120).
	if conv.IsParticipant(42, 0, true) {
		t.Error("an admin must not be a participant of a chef thread")
	}
}

func TestConversation_IsParticipant_Support(t *testing.T) {
	// A support thread: user 10 <-> the platform, no kitchen.
	conv := &domain.Conversation{Kind: domain.ConversationKindSupport, UserID: 10}

	// The target user.
	if !conv.IsParticipant(10, 0, false) {
		t.Error("the target user should be a participant of their support thread")
	}
	// Any admin, regardless of user id.
	if !conv.IsParticipant(999, 0, true) {
		t.Error("an admin should be a participant of a support thread")
	}
	// A non-admin, non-target user — even one who happens to be a chef — is not.
	if conv.IsParticipant(42, 3, false) {
		t.Error("an unrelated user must not access a support thread")
	}
}
