package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLoginSuccess(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"password": testPassword})
	res, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", res.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := got["token"]; ok {
		t.Fatal("token must not appear in login response body")
	}

	var found bool
	for _, c := range res.Cookies() {
		if c.Name == "token" {
			found = true
			if !c.HttpOnly {
				t.Error("token cookie must be HttpOnly")
			}
			if c.Value == "" {
				t.Error("token cookie must not be empty")
			}
		}
	}
	if !found {
		t.Fatal("no token cookie in response")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"password": "wrongpassword"})
	res, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", res.StatusCode)
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	res := do(t, http.MethodGet, "/clients", nil, nil)
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", res.StatusCode)
	}
}

func TestAuthMe(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodGet, "/auth/me", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("got %d, want 204", res.StatusCode)
	}
}

func TestLogout(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodPost, "/auth/logout", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: got %d, want 204", res.StatusCode)
	}

	// Response must instruct the browser to delete the cookie (MaxAge < 0).
	// JWTs are stateless so the token itself stays valid until expiry —
	// logout works by clearing the cookie on the client side.
	var found bool
	for _, c := range res.Cookies() {
		if c.Name == "token" {
			found = true
			if c.MaxAge >= 0 {
				t.Errorf("token cookie MaxAge = %d, want < 0 (delete)", c.MaxAge)
			}
		}
	}
	if !found {
		t.Error("logout response did not set a token cookie to clear it")
	}
}

func TestSetupRequiredReturnsFalse(t *testing.T) {
	// TestMain seeds a password hash so setup should not be required.
	res, err := http.Get(srv.URL + "/auth/setup-required")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", res.StatusCode)
	}

	var got map[string]bool
	decodeJSON(t, res, &got)
	if got["required"] {
		t.Fatal("required = true, want false (password is already set)")
	}
}

func TestSetupBlockedWhenAlreadySet(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"setup_secret": testSetupSecret, "password": "newpassword12345"})
	res, err := http.Post(srv.URL+"/auth/setup", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d, want 403 (setup already done)", res.StatusCode)
	}
}

func TestSetupWeakPasswordRejected(t *testing.T) {
	// Clear password hash so setup is available, run the test, then restore.
	_, err := pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = ''`)
	if err != nil {
		t.Fatalf("clear hash: %v", err)
	}
	t.Cleanup(func() {
		hash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
		pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = $1`, string(hash))
	})

	body, _ := json.Marshal(map[string]string{"setup_secret": testSetupSecret, "password": "short"})
	res, err := http.Post(srv.URL+"/auth/setup", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400 (weak password)", res.StatusCode)
	}
}

func TestSetupInvalidSecretRejected(t *testing.T) {
	// Clear password hash so setup is available, run the test, then restore.
	_, err := pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = ''`)
	if err != nil {
		t.Fatalf("clear hash: %v", err)
	}
	t.Cleanup(func() {
		hash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
		pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = $1`, string(hash))
	})

	body, _ := json.Marshal(map[string]string{
		"setup_secret": "wrong-setup-secret",
		"password":     "newpassword12345",
	})
	res, err := http.Post(srv.URL+"/auth/setup", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401 (invalid setup secret)", res.StatusCode)
	}
}

func TestSetupFlow(t *testing.T) {
	const setupPassword = "brandnewpassword99"

	// Clear password hash so setup is required.
	_, err := pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = ''`)
	if err != nil {
		t.Fatalf("clear hash: %v", err)
	}
	t.Cleanup(func() {
		hash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
		pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = $1`, string(hash))
	})

	// setup-required must now return true.
	res, err := http.Get(srv.URL + "/auth/setup-required")
	if err != nil {
		t.Fatalf("setup-required: %v", err)
	}
	var sr map[string]bool
	decodeJSON(t, res, &sr)
	if !sr["required"] {
		t.Fatal("required = false, want true after clearing hash")
	}

	// POST /auth/setup must accept the new password.
	body, _ := json.Marshal(map[string]string{"setup_secret": testSetupSecret, "password": setupPassword})
	res, err = http.Post(srv.URL+"/auth/setup", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("setup: got %d, want 204", res.StatusCode)
	}

	// A second setup call must be rejected (only one setup allowed).
	body, _ = json.Marshal(map[string]string{"setup_secret": testSetupSecret, "password": setupPassword})
	res, err = http.Post(srv.URL+"/auth/setup", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("second setup: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("second setup: got %d, want 403", res.StatusCode)
	}

	// Login with the newly set password must succeed.
	loginBody, _ := json.Marshal(map[string]string{"password": setupPassword})
	loginRes, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login after setup: %v", err)
	}
	loginRes.Body.Close()
	if loginRes.StatusCode != http.StatusOK {
		t.Fatalf("login after setup: got %d, want 200", loginRes.StatusCode)
	}
}

func TestChangePassword(t *testing.T) {
	const newPassword = "changedpassword99"
	cookie := login(t)

	// Change the password.
	res := do(t, http.MethodPost, "/auth/password", map[string]string{
		"current_password": testPassword,
		"password":         newPassword,
	}, cookie)
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("change password: got %d, want 204", res.StatusCode)
	}

	// Old password must no longer work.
	badBody, _ := json.Marshal(map[string]string{"password": testPassword})
	badRes, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(badBody))
	if err != nil {
		t.Fatalf("login with old password: %v", err)
	}
	badRes.Body.Close()
	if badRes.StatusCode != http.StatusUnauthorized {
		t.Fatalf("old password: got %d, want 401", badRes.StatusCode)
	}

	// Restore original password so other tests keep working.
	loginNew, _ := json.Marshal(map[string]string{"password": newPassword})
	loginRes, _ := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginNew))
	var newCookie *http.Cookie
	for _, c := range loginRes.Cookies() {
		if c.Name == "token" {
			newCookie = c
		}
	}
	loginRes.Body.Close()

	restore := do(t, http.MethodPost, "/auth/password", map[string]string{
		"current_password": newPassword,
		"password":         testPassword,
	}, newCookie)
	restore.Body.Close()
	if restore.StatusCode != http.StatusNoContent {
		t.Fatalf("restore password: got %d, want 204", restore.StatusCode)
	}
}

func TestChangePasswordWrongCurrentPassword(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodPost, "/auth/password", map[string]string{
		"current_password": "thisisthewrongpassword",
		"password":         "brandnewpassword99",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401 (wrong current password)", res.StatusCode)
	}
}

func TestChangePasswordTooShort(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodPost, "/auth/password", map[string]string{
		"current_password": testPassword,
		"password":         "short",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400 (new password too short)", res.StatusCode)
	}
}
