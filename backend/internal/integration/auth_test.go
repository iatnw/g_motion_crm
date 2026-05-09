package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
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

	// Cookie should be cleared (MaxAge < 0 or Expires in the past)
	for _, c := range res.Cookies() {
		if c.Name == "token" && c.MaxAge >= 0 {
			t.Error("token cookie should be cleared after logout")
		}
	}

	// Subsequent request with old cookie must be rejected
	res2 := do(t, http.MethodGet, "/auth/me", nil, cookie)
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("after logout: got %d, want 401", res2.StatusCode)
	}
}
