package domain_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func TestConversation_IsParticipant(t *testing.T) {
	conv := &domain.Conversation{UserID: 10, ChefID: 3}

	// The customer (by user id).
	if !conv.IsParticipant(10, 0) {
		t.Error("customer should be a participant")
	}
	// The owning chef (by chef id), even though their user id differs.
	if !conv.IsParticipant(99, 3) {
		t.Error("owning chef should be a participant")
	}
	// An unrelated user with no matching chef id.
	if conv.IsParticipant(42, 0) {
		t.Error("unrelated user should not be a participant")
	}
	// A different chef.
	if conv.IsParticipant(42, 7) {
		t.Error("a different chef should not be a participant")
	}
}
