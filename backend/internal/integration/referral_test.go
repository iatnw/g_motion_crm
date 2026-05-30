package integration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func createDealForReferral(t *testing.T, cookie *http.Cookie) string {
	t.Helper()
	unique := uuid.NewString()[:8]
	clientRes := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "Ref",
		"last_name":  "Client",
		"phone":      "555-" + unique,
		"email":      "ref-" + unique + "@example.com",
	}, cookie)
	if clientRes.StatusCode != http.StatusCreated {
		t.Fatalf("create client: got %d, want 201", clientRes.StatusCode)
	}
	var client map[string]any
	decodeJSON(t, clientRes, &client)

	dealRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": client["id"],
	}, cookie)
	if dealRes.StatusCode != http.StatusCreated {
		t.Fatalf("create deal: got %d, want 201", dealRes.StatusCode)
	}
	var deal map[string]any
	decodeJSON(t, dealRes, &deal)
	return deal["id"].(string)
}

func TestListAllReferrals(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	// Create two referrals on the same deal
	res1 := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Alice",
		"last_name":  "Smith",
		"phone":      "555-1111",
		"amount":     200,
	}, cookie)
	res1.Body.Close()
	if res1.StatusCode != http.StatusCreated {
		t.Fatalf("create referral 1: got %d, want 201", res1.StatusCode)
	}

	res2 := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Bob",
		"last_name":  "Jones",
		"phone":      "555-2222",
		"amount":     300,
	}, cookie)
	res2.Body.Close()
	if res2.StatusCode != http.StatusCreated {
		t.Fatalf("create referral 2: got %d, want 201", res2.StatusCode)
	}

	// GET /referrals must return both with deal_number
	listRes := do(t, http.MethodGet, "/referrals", nil, cookie)
	var got []map[string]any
	decodeJSON(t, listRes, &got)

	if len(got) != 2 {
		t.Fatalf("expected 2 referrals, got %d", len(got))
	}
	for _, r := range got {
		if r["deal_number"] == nil || r["deal_number"] == "" {
			t.Errorf("referral missing deal_number: %v", r)
		}
	}
}

func TestListAllReferralsRequiresAuth(t *testing.T) {
	res := do(t, http.MethodGet, "/referrals", nil, nil)
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", res.StatusCode)
	}
}

func TestListReferralsByDealIDs(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)
	otherDealID := createDealForReferral(t, cookie)

	for _, item := range []struct {
		dealID string
		name   string
	}{
		{dealID: dealID, name: "Batch"},
		{dealID: otherDealID, name: "Other"},
	} {
		res := do(t, http.MethodPost, "/deals/"+item.dealID+"/referrals", map[string]any{
			"first_name": item.name,
			"amount":     100,
		}, cookie)
		res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create referral: got %d, want 201", res.StatusCode)
		}
	}

	listRes := do(t, http.MethodGet, "/referrals?deal_ids="+dealID, nil, cookie)
	var got []map[string]any
	decodeJSON(t, listRes, &got)

	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", listRes.StatusCode)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 referral, got %d", len(got))
	}
	if got[0]["deal_id"] != dealID {
		t.Errorf("deal_id = %v, want %s", got[0]["deal_id"], dealID)
	}
}

func TestCreateReferralManually(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	res := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Carol",
		"last_name":  "White",
		"phone":      "555-3333",
		"amount":     500,
	}, cookie)
	var ref map[string]any
	decodeJSON(t, res, &ref)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", res.StatusCode)
	}
	if ref["first_name"] != "Carol" {
		t.Errorf("first_name = %v, want Carol", ref["first_name"])
	}
	if ref["deal_id"] != dealID {
		t.Errorf("deal_id = %v, want %s", ref["deal_id"], dealID)
	}
	if ref["status"] != "none" {
		t.Errorf("status = %v, want none", ref["status"])
	}
}

func TestCreateReferralWithStatus(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	res := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Dana",
		"last_name":  "Paid",
		"amount":     250,
		"status":     "paid",
	}, cookie)
	var ref map[string]any
	decodeJSON(t, res, &ref)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", res.StatusCode)
	}
	if ref["status"] != "paid" {
		t.Errorf("status = %v, want paid", ref["status"])
	}
}

func TestCreateReferralRejectsInvalidStatus(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	res := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Evan",
		"status":     "unknown",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}

func TestCreateReferralMissingName(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	res := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "",
		"last_name":  "",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}

func TestDeleteReferral(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)

	createRes := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Dave",
		"last_name":  "Brown",
	}, cookie)
	var ref map[string]any
	decodeJSON(t, createRes, &ref)

	delRes := do(t, http.MethodDelete, "/referrals/"+ref["id"].(string), nil, cookie)
	defer delRes.Body.Close()
	if delRes.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: got %d, want 204", delRes.StatusCode)
	}

	// Must not appear in the list anymore
	listRes := do(t, http.MethodGet, "/referrals", nil, cookie)
	var got []map[string]any
	decodeJSON(t, listRes, &got)
	if len(got) != 0 {
		t.Fatalf("expected 0 referrals after delete, got %d", len(got))
	}
}

func TestUpdateReferralAmount(t *testing.T) {
	t.Cleanup(truncateData)
	cookie := login(t)

	dealID := createDealForReferral(t, cookie)
	dealRes := do(t, http.MethodPut, "/deals/"+dealID, map[string]any{
		"client_id":                referralDealClientID(t, dealID),
		"broker_fee":               1000,
		"broker_commission":        200,
		"deal_status":              "pending",
		"broker_fee_status":        "pending",
		"broker_commission_status": "pending",
	}, cookie)
	dealRes.Body.Close()
	if dealRes.StatusCode != http.StatusOK {
		t.Fatalf("update deal financials: got %d, want 200", dealRes.StatusCode)
	}

	createRes := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Amount",
		"last_name":  "Update",
		"amount":     100,
	}, cookie)
	var created map[string]any
	decodeJSON(t, createRes, &created)
	if createRes.StatusCode != http.StatusCreated {
		t.Fatalf("create referral: got %d, want 201", createRes.StatusCode)
	}

	updateRes := do(t, http.MethodPatch, "/referrals/"+created["id"].(string)+"/amount", map[string]any{
		"amount": 250,
	}, cookie)
	var updated map[string]any
	decodeJSON(t, updateRes, &updated)
	if updateRes.StatusCode != http.StatusOK {
		t.Fatalf("update referral amount: got %d, want 200", updateRes.StatusCode)
	}
	if updated["amount"] != float64(250) {
		t.Fatalf("amount = %v, want 250", updated["amount"])
	}

	var net float64
	if err := pool.QueryRow(context.Background(), `SELECT your_net FROM deals WHERE id = $1`, dealID).Scan(&net); err != nil {
		t.Fatalf("query deal net: %v", err)
	}
	if net != 550 {
		t.Fatalf("deal net = %v, want 550", net)
	}
}

func referralDealClientID(t *testing.T, dealID string) string {
	t.Helper()
	var clientID string
	if err := pool.QueryRow(context.Background(), `SELECT client_id::text FROM deals WHERE id = $1`, dealID).Scan(&clientID); err != nil {
		t.Fatalf("query deal client: %v", err)
	}
	return clientID
}
