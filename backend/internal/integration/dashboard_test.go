package integration_test

import (
	"net/http"
	"testing"
)

func TestDashboardDealerAverageIsNullWithoutPaidCheck(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	dealerRes := do(t, http.MethodPost, "/dealers", map[string]any{
		"name": "No Paid Checks Dealer",
	}, cookie)
	var dealer map[string]any
	decodeJSON(t, dealerRes, &dealer)
	if dealerRes.StatusCode != http.StatusCreated {
		t.Fatalf("create dealer: got %d, want 201", dealerRes.StatusCode)
	}

	clientID := createTestClient(t, cookie, "555-1013", "dashboard-dealer@example.com")
	dealRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"dealer_id": dealer["id"],
		"vin":       "1HGCV1F30JA000003",
	}, cookie)
	dealRes.Body.Close()
	if dealRes.StatusCode != http.StatusCreated {
		t.Fatalf("create deal: got %d, want 201", dealRes.StatusCode)
	}

	summaryRes := do(t, http.MethodGet, "/dashboard/summary", nil, cookie)
	var summary map[string]any
	decodeJSON(t, summaryRes, &summary)
	if summaryRes.StatusCode != http.StatusOK {
		t.Fatalf("dashboard summary: got %d, want 200", summaryRes.StatusCode)
	}
	dealers := summary["dealers"].([]any)
	if len(dealers) != 1 {
		t.Fatalf("dealers = %d, want 1", len(dealers))
	}
	dealerStat := dealers[0].(map[string]any)
	if dealerStat["avg_days_to_payment"] != nil {
		t.Fatalf("avg_days_to_payment = %v, want null", dealerStat["avg_days_to_payment"])
	}
}
