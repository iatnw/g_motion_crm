package integration_test

import (
	"net/http"
	"testing"
)

func TestCreateClient(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	res := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "Alice",
		"last_name":  "Smith",
		"phone":      "555-0001",
		"email":      "alice@example.com",
		"status":     "active",
	}, cookie)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", res.StatusCode)
	}

	var got map[string]any
	decodeJSON(t, res, &got)

	if got["id"] == nil {
		t.Fatal("response missing id")
	}
	if got["first_name"] != "Alice" {
		t.Fatalf("first_name = %v, want Alice", got["first_name"])
	}
	if got["phone"] != "555-0001" {
		t.Fatalf("phone = %v, want 555-0001", got["phone"])
	}
}

func TestGetClient(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	// Create
	res := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "Bob",
		"last_name":  "Jones",
		"phone":      "555-0002",
		"email":      "bob@example.com",
		"status":     "active",
	}, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d, want 201", res.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, res, &created)
	id := created["id"].(string)

	// Get
	res2 := do(t, http.MethodGet, "/clients/"+id, nil, cookie)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("get: got %d, want 200", res2.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res2, &got)
	if got["id"] != id {
		t.Fatalf("id = %v, want %s", got["id"], id)
	}
	if got["last_name"] != "Jones" {
		t.Fatalf("last_name = %v, want Jones", got["last_name"])
	}
}

func TestListClients(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	// Create two clients
	for i, phone := range []string{"555-0010", "555-0011"} {
		res := do(t, http.MethodPost, "/clients", map[string]any{
			"first_name": "Client",
			"last_name":  "Test",
			"phone":      phone,
			"email":      phone + "@example.com",
			"status":     "active",
		}, cookie)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create client %d: got %d", i, res.StatusCode)
		}
		res.Body.Close()
	}

	res := do(t, http.MethodGet, "/clients", nil, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list: got %d, want 200", res.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res, &got)

	data, ok := got["data"].([]any)
	if !ok || len(data) < 2 {
		t.Fatalf("expected at least 2 clients, got %v", got["data"])
	}
}

func TestUpdateClient(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	// Create
	res := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "Carol",
		"last_name":  "White",
		"phone":      "555-0020",
		"email":      "carol@example.com",
		"status":     "active",
	}, cookie)
	var created map[string]any
	decodeJSON(t, res, &created)
	id := created["id"].(string)

	// Update
	res2 := do(t, http.MethodPut, "/clients/"+id, map[string]any{
		"first_name": "Caroline",
		"last_name":  "White",
		"phone":      "555-0020",
		"email":      "caroline@example.com",
		"status":     "active",
	}, cookie)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("update: got %d, want 200", res2.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res2, &got)
	if got["first_name"] != "Caroline" {
		t.Fatalf("first_name = %v, want Caroline", got["first_name"])
	}
}

func TestGetClientNotFound(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodGet, "/clients/00000000-0000-0000-0000-000000000000", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("got %d, want 404", res.StatusCode)
	}
}

func TestCreateClientMissingFields(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "OnlyFirst",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}
