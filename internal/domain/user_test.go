package domain

import "testing"

func TestValidRole(t *testing.T) {
	cases := map[string]bool{
		RoleCustomer: true,
		RoleChef:     true,
		RoleAdmin:    true,
		"":           false,
		"superuser":  false,
	}
	for role, want := range cases {
		if got := ValidRole(role); got != want {
			t.Errorf("ValidRole(%q) = %v, want %v", role, got, want)
		}
	}
}

func TestNewUserDefaults(t *testing.T) {
	u := NewUser("yasin", "yasin@example.com", "hash")

	if u.Role != RoleCustomer {
		t.Errorf("default role = %q, want %q", u.Role, RoleCustomer)
	}
	if !u.IsActive {
		t.Error("new user should be active")
	}
	if u.IsVerified {
		t.Error("new user should not be verified")
	}
	if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Error("timestamps should be set")
	}
}

func TestRoleHelpers(t *testing.T) {
	if !(&User{Role: RoleChef}).IsChef() {
		t.Error("IsChef should be true for chef")
	}
	if !(&User{Role: RoleAdmin}).IsAdmin() {
		t.Error("IsAdmin should be true for admin")
	}
	if (&User{Role: RoleCustomer}).IsChef() {
		t.Error("customer is not a chef")
	}
}
