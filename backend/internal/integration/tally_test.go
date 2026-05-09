package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func tallyPayload(eventID, submissionID, phone, email, name string) []byte {
	payload := map[string]any{
		"eventId":   eventID,
		"eventType": "FORM_RESPONSE",
		"createdAt": time.Now().Format(time.RFC3339),
		"data": map[string]any{
			"responseId":   "resp-" + eventID,
			"submissionId": submissionID,
			"formId":       "form-1",
			"formName":     "Credit App",
			"createdAt":    time.Now().Format(time.RFC3339),
			"fields": []map[string]any{
				{"key": "q1", "label": "Full legal name", "type": "INPUT_TEXT", "value": name},
				{"key": "q2", "label": "Phone Number", "type": "INPUT_PHONE_NUMBER", "value": phone},
				{"key": "q3", "label": "Email", "type": "INPUT_EMAIL", "value": email},
			},
		},
	}
	b, _ := json.Marshal(payload)
	return b
}

func postTally(t *testing.T, body []byte) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/webhooks/tally", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tally-signature", tallySignature(body))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("tally request: %v", err)
	}
	return res
}

func TestTallyWebhookInvalidSignature(t *testing.T) {
	body := tallyPayload("evt-bad", "sub-bad", "555-8001", "bad@example.com", "Bad Actor")

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/webhooks/tally", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tally-signature", "invalidsignature")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", res.StatusCode)
	}
}

func TestTallyWebhookCreatesClient(t *testing.T) {
	t.Cleanup(truncateData)

	body := tallyPayload("evt-001", "sub-001", "555-8002", "tally1@example.com", "Jane Doe")
	res := postTally(t, body)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", res.StatusCode)
	}

	// Verify client was created in the database
	cookie := login(t)
	listRes := do(t, http.MethodGet, "/clients", nil, cookie)
	var got map[string]any
	decodeJSON(t, listRes, &got)

	data := got["data"].([]any)
	if len(data) == 0 {
		t.Fatal("expected client to be created by tally webhook")
	}

	client := data[0].(map[string]any)
	if client["first_name"] != "Jane" {
		t.Fatalf("first_name = %v, want Jane", client["first_name"])
	}
}

func TestTallyWebhookDeduplication(t *testing.T) {
	t.Cleanup(truncateData)

	body := tallyPayload("evt-002", "sub-002", "555-8003", "tally2@example.com", "John Smith")

	// First delivery
	res1 := postTally(t, body)
	defer res1.Body.Close()
	if res1.StatusCode != http.StatusOK {
		t.Fatalf("first delivery: got %d, want 200", res1.StatusCode)
	}

	// Second delivery — same event, should be a no-op
	res2 := postTally(t, body)
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("second delivery: got %d, want 200", res2.StatusCode)
	}

	// Only one client should exist
	cookie := login(t)
	listRes := do(t, http.MethodGet, "/clients", nil, cookie)
	var got map[string]any
	decodeJSON(t, listRes, &got)

	// Check tally_webhook_events table directly for the processed_at column
	var count int
	err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM tally_webhook_events WHERE event_id = 'evt-002'`,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 webhook event row, got %d", count)
	}

	data := got["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected exactly 1 client (no duplicates), got %d", len(data))
	}
}
