package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDealReferences(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)
	clientID := createTestClient(t, cookie, "555-8100", "references@example.com")
	res := do(t, http.MethodPost, "/deals", map[string]any{"client_id": clientID, "notes": "original"}, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create without broker/dealer: %d", res.StatusCode)
	}
	var original map[string]any
	decodeJSON(t, res, &original)
	dealID := original["id"].(string)

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, field := range []string{"client_id", "broker_id", "dealer_id"} {
			t.Run(method+"/"+field, func(t *testing.T) {
				body := map[string]any{"client_id": clientID, "notes": "should not persist", field: uuid.NewString()}
				path := "/deals"
				if method == http.MethodPut {
					path += "/" + dealID
				}
				res := do(t, method, path, body, cookie)
				defer res.Body.Close()
				message, _ := io.ReadAll(res.Body)
				if res.StatusCode != http.StatusBadRequest {
					t.Fatalf("status %d, body %s", res.StatusCode, message)
				}
				if strings.TrimSpace(string(message)) != field+" does not reference an existing record" {
					t.Fatalf("unexpected message: %s", message)
				}
			})
		}
	}
	res = do(t, http.MethodGet, "/deals/"+dealID, nil, cookie)
	var saved map[string]any
	decodeJSON(t, res, &saved)
	if saved["notes"] != "original" || saved["client_id"] != clientID {
		t.Fatalf("failed writes changed deal: %v", saved)
	}
	res = do(t, http.MethodGet, "/deals", nil, cookie)
	var list struct {
		Data []map[string]any `json:"data"`
	}
	decodeJSON(t, res, &list)
	if len(list.Data) != 1 {
		t.Fatalf("failed creates persisted deals: %d", len(list.Data))
	}
}
