package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// registerCustomer registers a plain customer account and returns its token.
func registerCustomer(t *testing.T, srv http.Handler, username, email string) string {
	t.Helper()
	body := `{"username":"` + username + `","email":"` + email + `","password":"secret123"}`
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/register", "", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup register failed: %d (%s)", rec.Code, rec.Body)
	}
	var reg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &reg); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	return reg.Token
}

func TestProfileHTTP_RequiresAuth(t *testing.T) {
	srv := newTestServer()
	for _, tc := range []struct{ method, path string }{
		{http.MethodPut, "/api/v2/users/me"},
		{http.MethodPut, "/api/v2/auth/password"},
		{http.MethodPut, "/api/v2/chefs/me"},
	} {
		if rec := do(t, srv, tc.method, tc.path, "", `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without token = %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}

func TestProfileHTTP_UpdateOwnProfile(t *testing.T) {
	srv := newTestServer()
	token := registerCustomer(t, srv, "cust", "cust@example.com")

	rec := do(t, srv, http.MethodPut, "/api/v2/users/me", token,
		`{"phone_number":"+90 555 111 22 33","city":"Istanbul","latitude":41.0,"longitude":29.0}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update profile = %d (%s)", rec.Code, rec.Body)
	}
	var user map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &user)
	if user["phone_number"] != "+90 555 111 22 33" || user["city"] != "Istanbul" {
		t.Errorf("profile not applied: %v", user)
	}
	if _, leaked := user["password_hash"]; leaked {
		t.Error("password_hash leaked in the response")
	}
	// Identity fields silently ignored even if sent.
	rec = do(t, srv, http.MethodPut, "/api/v2/users/me", token,
		`{"email":"evil@example.com","role":"admin","city":"Ankara"}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &user)
	if user["email"] != "cust@example.com" || user["role"] != "customer" {
		t.Errorf("identity fields must be immutable via profile update: %v", user)
	}

	// lat without lng -> 400.
	if rec := do(t, srv, http.MethodPut, "/api/v2/users/me", token, `{"latitude":41.0}`); rec.Code != http.StatusBadRequest {
		t.Errorf("lat-only update = %d, want 400", rec.Code)
	}

	// Email-notification preference: defaults on, toggles off, omitted keeps.
	rec = do(t, srv, http.MethodGet, "/api/v2/auth/me", token, "")
	_ = json.Unmarshal(rec.Body.Bytes(), &user)
	if user["email_notifications"] != true {
		t.Errorf("email_notifications default = %v, want true", user["email_notifications"])
	}
	rec = do(t, srv, http.MethodPut, "/api/v2/users/me", token, `{"email_notifications":false}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &user)
	if user["email_notifications"] != false {
		t.Errorf("opt-out not applied: %v", user["email_notifications"])
	}
	rec = do(t, srv, http.MethodPut, "/api/v2/users/me", token, `{"city":"Ankara"}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &user)
	if user["email_notifications"] != false {
		t.Errorf("omitted field must keep the opt-out: %v", user["email_notifications"])
	}
}

func TestProfileHTTP_ChangePassword(t *testing.T) {
	srv := newTestServer()
	token := registerCustomer(t, srv, "cust", "cust@example.com")

	// Wrong current password -> 401.
	rec := do(t, srv, http.MethodPut, "/api/v2/auth/password", token,
		`{"current_password":"wrong","new_password":"newpass1"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong current = %d, want 401", rec.Code)
	}

	// Success, then old password fails and the new one logs in.
	rec = do(t, srv, http.MethodPut, "/api/v2/auth/password", token,
		`{"current_password":"secret123","new_password":"newpass1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("change password = %d (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"cust@example.com","password":"secret123"}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("old password login = %d, want 401", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"cust@example.com","password":"newpass1"}`); rec.Code != http.StatusOK {
		t.Errorf("new password login = %d, want 200", rec.Code)
	}
}

func TestProfileHTTP_UpdateChefProfile(t *testing.T) {
	srv := newTestServer()

	// Customers don't get the chef endpoint.
	custToken := registerCustomer(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me", custToken, `{}`); rec.Code != http.StatusForbidden {
		t.Errorf("customer on PUT /chefs/me = %d, want 403", rec.Code)
	}

	chefToken := registerAndToken(t, srv, "chef", "chef@example.com")

	// No profile yet -> 404 (same signal the onboarding flow uses).
	rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me", chefToken,
		`{"business_name":"New Name","kitchen_address":"1 Main St"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("update without profile = %d, want 404 (%s)", rec.Code, rec.Body)
	}

	// Create then update.
	rec = do(t, srv, http.MethodPost, "/api/v2/chefs", chefToken,
		`{"business_name":"Old Name","kitchen_address":"1 Main St"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create profile = %d (%s)", rec.Code, rec.Body)
	}
	rec = do(t, srv, http.MethodPut, "/api/v2/chefs/me", chefToken,
		`{"business_name":"New Name","kitchen_address":"2 Side St","specialty":"soups","delivery_radius":9}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update profile = %d (%s)", rec.Code, rec.Body)
	}
	var chef map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &chef)
	if chef["business_name"] != "New Name" || chef["kitchen_address"] != "2 Side St" || chef["delivery_radius"] != float64(9) {
		t.Errorf("chef update not applied: %v", chef)
	}

	// Missing business name -> 400.
	rec = do(t, srv, http.MethodPut, "/api/v2/chefs/me", chefToken, `{"kitchen_address":"2 Side St"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing business_name = %d, want 400", rec.Code)
	}
}
