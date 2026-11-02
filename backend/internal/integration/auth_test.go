package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"net/http"
	"testing"
	"time"

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
	otherCookie := login(t)

	res := do(t, http.MethodPost, "/auth/logout", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: got %d, want 204", res.StatusCode)
	}

	// Response must instruct the browser to delete the cookie (MaxAge < 0).
	// The original token must also be rejected by the server.
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
	for _, oldCookie := range []*http.Cookie{cookie, otherCookie} {
		replay := do(t, http.MethodGet, "/auth/me", nil, oldCookie)
		replay.Body.Close()
		if replay.StatusCode != http.StatusUnauthorized {
			t.Fatalf("revoked token: got %d, want 401", replay.StatusCode)
		}
	}
	fresh := do(t, http.MethodGet, "/auth/me", nil, login(t))
	fresh.Body.Close()
	if fresh.StatusCode != http.StatusNoContent {
		t.Fatalf("fresh login: got %d, want 204", fresh.StatusCode)
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
	otherCookie := login(t)
	t.Cleanup(func() {
		hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
		if err != nil {
			t.Error(err)
			return
		}
		if _, err := pool.Exec(context.Background(), `UPDATE settings SET owner_password_hash = $1`, string(hash)); err != nil {
			t.Error(err)
		}
	})

	// Change the password.
	res := do(t, http.MethodPost, "/auth/password", map[string]string{
		"current_password": testPassword,
		"password":         newPassword,
	}, cookie)
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("change password: got %d, want 204", res.StatusCode)
	}

	for _, oldCookie := range []*http.Cookie{cookie, otherCookie} {
		replay := do(t, http.MethodGet, "/auth/me", nil, oldCookie)
		replay.Body.Close()
		if replay.StatusCode != http.StatusUnauthorized {
			t.Fatalf("revoked token: got %d, want 401", replay.StatusCode)
		}
	}
	cleared := false
	for _, c := range res.Cookies() {
		if c.Name == "token" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("password change did not clear cookie")
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
	loginRes, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginNew))
	if err != nil {
		t.Fatal(err)
	}
	if loginRes.StatusCode != http.StatusOK {
		loginRes.Body.Close()
		t.Fatalf("new password login: got %d, want 200", loginRes.StatusCode)
	}
	var newCookie *http.Cookie
	for _, c := range loginRes.Cookies() {
		if c.Name == "token" {
			newCookie = c
		}
	}
	loginRes.Body.Close()

	fresh := do(t, http.MethodGet, "/auth/me", nil, newCookie)
	fresh.Body.Close()
	if fresh.StatusCode != http.StatusNoContent {
		t.Fatalf("new session: got %d, want 204", fresh.StatusCode)
	}
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
	stillValid := do(t, http.MethodGet, "/auth/me", nil, cookie)
	stillValid.Body.Close()
	if stillValid.StatusCode != http.StatusNoContent {
		t.Fatalf("session after rejected change: got %d, want 204", stillValid.StatusCode)
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

func TestLogoutRevokesPendingTOTPToken(t *testing.T) {
	var version int64
	if err := pool.QueryRow(context.Background(), `SELECT session_version FROM settings WHERE id = 1`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "owner", "stage": "totp_pending", "session_version": version,
		"exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte("test-jwt-secret-for-integration-tests-only"))
	if err != nil {
		t.Fatal(err)
	}
	res := do(t, http.MethodPost, "/auth/logout", nil, login(t))
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: got %d, want 204", res.StatusCode)
	}
	verify := do(t, http.MethodPost, "/auth/totp/verify", map[string]string{"code": "123456"}, &http.Cookie{Name: "token", Value: token})
	verify.Body.Close()
	if verify.StatusCode != http.StatusUnauthorized {
		t.Fatalf("revoked pending token: got %d, want 401", verify.StatusCode)
	}
}

func TestDisableTOTPReauthentication(t *testing.T) {
	const secret = "JBSWY3DPEHPK3PXP"
	cookie := login(t)
	if _, err := pool.Exec(context.Background(), `UPDATE settings SET owner_totp_secret = $1`, secret); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `UPDATE settings SET owner_totp_secret = NULL`); err != nil {
			t.Error(err)
		}
	})
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		body any
		want int
	}{
		{"missing body", nil, 400},
		{"missing credentials", map[string]string{}, 400},
		{"wrong password", map[string]string{"current_password": "wrong", "code": code}, 401},
		{"wrong code", map[string]string{"current_password": testPassword, "code": "invalid"}, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := do(t, http.MethodPost, "/auth/totp/disable", tc.body, cookie)
			res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("got %d, want %d", res.StatusCode, tc.want)
			}
			var actual string
			if err := pool.QueryRow(context.Background(), `SELECT owner_totp_secret FROM settings WHERE id = 1`).Scan(&actual); err != nil {
				t.Fatal(err)
			}
			if actual != secret {
				t.Fatal("failed reauthentication changed TOTP secret")
			}
		})
	}
	replacement := do(t, http.MethodPost, "/auth/totp/enable", map[string]string{"secret": secret, "code": code}, cookie)
	replacement.Body.Close()
	if replacement.StatusCode != http.StatusConflict {
		t.Fatalf("replace enabled TOTP: got %d, want 409", replacement.StatusCode)
	}
	res := do(t, http.MethodPost, "/auth/totp/disable", map[string]string{"current_password": testPassword, "code": code}, cookie)
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("valid disable: got %d, want 204", res.StatusCode)
	}
	var disabled bool
	if err := pool.QueryRow(context.Background(), `SELECT owner_totp_secret IS NULL FROM settings WHERE id = 1`).Scan(&disabled); err != nil {
		t.Fatal(err)
	}
	if !disabled {
		t.Fatal("TOTP still enabled")
	}
}
