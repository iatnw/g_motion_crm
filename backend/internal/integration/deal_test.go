package integration_test

import (
	"net/http"
	"testing"
)

// createTestClient is a helper that creates a client and returns its ID.
func createTestClient(t *testing.T, cookie *http.Cookie, phone, email string) string {
	t.Helper()
	res := do(t, http.MethodPost, "/clients", map[string]any{
		"first_name": "Test",
		"last_name":  "Client",
		"phone":      phone,
		"email":      email,
		"status":     "active",
	}, cookie)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create client: got %d, want 201", res.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res, &got)
	return got["id"].(string)
}

func TestCreateDeal(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1001", "deal1@example.com")

	res := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"make":      "Toyota",
		"model":     "Camry",
	}, cookie)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", res.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res, &got)
	if got["id"] == nil {
		t.Fatal("response missing id")
	}
	if got["client_id"] != clientID {
		t.Fatalf("client_id = %v, want %s", got["client_id"], clientID)
	}
}

func TestGetDeal(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1002", "deal2@example.com")

	res := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"make":      "Honda",
		"model":     "Civic",
	}, cookie)
	var created map[string]any
	decodeJSON(t, res, &created)
	dealID := created["id"].(string)

	res2 := do(t, http.MethodGet, "/deals/"+dealID, nil, cookie)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("get: got %d, want 200", res2.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res2, &got)
	if got["id"] != dealID {
		t.Fatalf("id = %v, want %s", got["id"], dealID)
	}
}

func TestListDeals(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1003", "deal3@example.com")

	for _, model := range []string{"Accord", "Pilot"} {
		res := do(t, http.MethodPost, "/deals", map[string]any{
			"client_id": clientID,
			"make":      "Honda",
			"model":     model,
		}, cookie)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create deal %s: got %d", model, res.StatusCode)
		}
		res.Body.Close()
	}

	res := do(t, http.MethodGet, "/deals", nil, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list: got %d, want 200", res.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res, &got)
	data, ok := got["data"].([]any)
	if !ok || len(data) < 2 {
		t.Fatalf("expected at least 2 deals, got %v", got["data"])
	}
}

func TestListDealsByDealer(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1012", "dealer-deals@example.com")
	dealerRes := do(t, http.MethodPost, "/dealers", map[string]any{"name": "Dealer One"}, cookie)
	if dealerRes.StatusCode != http.StatusCreated {
		t.Fatalf("create dealer: got %d, want 201", dealerRes.StatusCode)
	}
	var dealer map[string]any
	decodeJSON(t, dealerRes, &dealer)

	for _, payload := range []map[string]any{
		{"client_id": clientID, "dealer_id": dealer["id"], "make": "Toyota"},
		{"client_id": clientID, "make": "Honda"},
	} {
		res := do(t, http.MethodPost, "/deals", payload, cookie)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create deal: got %d, want 201", res.StatusCode)
		}
		res.Body.Close()
	}

	assertDealListLength(t, cookie, "/deals?dealer_id="+dealer["id"].(string), 1)
}

func TestListDealsRejectsInvalidDealerID(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodGet, "/deals?dealer_id=not-a-uuid", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}

func TestListDealsRejectsInvalidStatusWithDealerID(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodGet, "/deals?dealer_id=00000000-0000-0000-0000-000000000000&status=unknown", nil, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}

func TestListDealsFiltersAwaitingVIN(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1010", "deal-vin@example.com")

	missingVINRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"make":      "Honda",
		"model":     "Pilot",
	}, cookie)
	if missingVINRes.StatusCode != http.StatusCreated {
		t.Fatalf("create missing VIN deal: got %d, want 201", missingVINRes.StatusCode)
	}
	var missingVINDeal map[string]any
	decodeJSON(t, missingVINRes, &missingVINDeal)

	presentVINRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"make":      "Honda",
		"model":     "Accord",
		"vin":       "1HGCV1F30JA000001",
	}, cookie)
	if presentVINRes.StatusCode != http.StatusCreated {
		t.Fatalf("create present VIN deal: got %d, want 201", presentVINRes.StatusCode)
	}
	presentVINRes.Body.Close()

	assertDealListLength(t, cookie, "/deals?vin=present", 1)
	assertDealListLength(t, cookie, "/deals?vin=missing", 1)

	updateRes := do(t, http.MethodPut, "/deals/"+missingVINDeal["id"].(string), map[string]any{
		"client_id":                clientID,
		"vin":                      "5FNYF6H50NB000001",
		"deal_status":              "pending",
		"broker_fee_status":        "pending",
		"broker_commission_status": "pending",
	}, cookie)
	if updateRes.StatusCode != http.StatusOK {
		t.Fatalf("update deal VIN: got %d, want 200", updateRes.StatusCode)
	}
	updateRes.Body.Close()

	assertDealListLength(t, cookie, "/deals?vin=present", 2)
	assertDealListLength(t, cookie, "/deals?vin=missing", 0)
}

func TestListDealsIncludesCancelledTab(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1011", "deal-cancelled@example.com")

	createRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
		"vin":       "1HGCV1F30JA000002",
	}, cookie)
	if createRes.StatusCode != http.StatusCreated {
		t.Fatalf("create deal: got %d, want 201", createRes.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, createRes, &created)
	dealID := created["id"].(string)

	cancelRes := do(t, http.MethodPatch, "/deals/"+dealID+"/status", map[string]any{
		"status": "cancelled",
	}, cookie)
	if cancelRes.StatusCode != http.StatusOK {
		t.Fatalf("cancel deal: got %d, want 200", cancelRes.StatusCode)
	}
	cancelRes.Body.Close()

	updateRes := do(t, http.MethodPut, "/deals/"+dealID, map[string]any{
		"client_id":                clientID,
		"deal_status":              "cancelled",
		"broker_fee_status":        "pending",
		"broker_commission_status": "pending",
	}, cookie)
	if updateRes.StatusCode != http.StatusOK {
		t.Fatalf("remove cancelled deal VIN: got %d, want 200", updateRes.StatusCode)
	}
	updateRes.Body.Close()

	assertDealListLength(t, cookie, "/deals?vin=present", 0)
	assertDealListLength(t, cookie, "/deals?status=cancelled", 1)
}

func assertDealListLength(t *testing.T, cookie *http.Cookie, path string, want int) {
	t.Helper()
	res := do(t, http.MethodGet, path, nil, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list deals: got %d, want 200", res.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res, &got)
	data, ok := got["data"].([]any)
	if !ok {
		t.Fatalf("expected deal data array, got %v", got["data"])
	}
	if len(data) != want {
		t.Fatalf("got %d deals, want %d", len(data), want)
	}
}

func TestUpdateDealStatus(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1004", "deal4@example.com")

	res := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
	}, cookie)
	var created map[string]any
	decodeJSON(t, res, &created)
	dealID := created["id"].(string)

	res2 := do(t, http.MethodPatch, "/deals/"+dealID+"/status", map[string]any{
		"status": "paid",
	}, cookie)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("update status: got %d, want 200", res2.StatusCode)
	}
	var got map[string]any
	decodeJSON(t, res2, &got)
	if got["deal_status"] != "paid" {
		t.Fatalf("deal_status = %v, want paid", got["deal_status"])
	}
	paidDate := got["broker_fee_paid_date"]
	if paidDate == nil {
		t.Fatal("broker_fee_paid_date missing after paid transition")
	}

	res3 := do(t, http.MethodPatch, "/deals/"+dealID+"/status", map[string]any{
		"status": "paid",
	}, cookie)
	var repeated map[string]any
	decodeJSON(t, res3, &repeated)
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("repeat paid status: got %d, want 200", res3.StatusCode)
	}
	if repeated["broker_fee_paid_date"] != paidDate {
		t.Fatalf("paid date changed on repeated paid status: got %v, want %v", repeated["broker_fee_paid_date"], paidDate)
	}
}

func TestFullDealUpdateSynchronizesPaymentDates(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1012", "deal-full-update@example.com")
	createRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
	}, cookie)
	var created map[string]any
	decodeJSON(t, createRes, &created)
	dealID := created["id"].(string)

	paidRes := do(t, http.MethodPut, "/deals/"+dealID, map[string]any{
		"client_id":                clientID,
		"deal_status":              "paid",
		"broker_fee_status":        "pending",
		"broker_commission_status": "paid",
	}, cookie)
	var paid map[string]any
	decodeJSON(t, paidRes, &paid)
	if paidRes.StatusCode != http.StatusOK {
		t.Fatalf("mark paid through full update: got %d, want 200", paidRes.StatusCode)
	}
	if paid["broker_fee_status"] != "paid" || paid["broker_fee_paid_date"] == nil {
		t.Fatalf("broker fee payment not synchronized: %v", paid)
	}
	if paid["broker_commission_paid_date"] == nil {
		t.Fatalf("broker commission paid date missing: %v", paid)
	}

	paidDate := paid["broker_fee_paid_date"]
	editRes := do(t, http.MethodPut, "/deals/"+dealID, map[string]any{
		"client_id":                clientID,
		"deal_status":              "paid",
		"broker_fee_status":        "paid",
		"broker_commission_status": "paid",
		"notes":                    "ordinary edit",
	}, cookie)
	var edited map[string]any
	decodeJSON(t, editRes, &edited)
	if editRes.StatusCode != http.StatusOK {
		t.Fatalf("edit paid deal: got %d, want 200", editRes.StatusCode)
	}
	if edited["broker_fee_paid_date"] != paidDate {
		t.Fatalf("paid date changed during ordinary edit: got %v, want %v", edited["broker_fee_paid_date"], paidDate)
	}

	pendingRes := do(t, http.MethodPut, "/deals/"+dealID, map[string]any{
		"client_id":                clientID,
		"deal_status":              "pending",
		"broker_fee_status":        "paid",
		"broker_commission_status": "pending",
	}, cookie)
	var pending map[string]any
	decodeJSON(t, pendingRes, &pending)
	if pendingRes.StatusCode != http.StatusOK {
		t.Fatalf("return deal to pending: got %d, want 200", pendingRes.StatusCode)
	}
	if pending["broker_fee_status"] != "pending" || pending["broker_fee_paid_date"] != nil {
		t.Fatalf("pending broker fee payment not synchronized: %v", pending)
	}
	if pending["broker_commission_paid_date"] != nil {
		t.Fatalf("pending broker commission retained paid date: %v", pending)
	}
}

func TestCreateDealMissingClientID(t *testing.T) {
	cookie := login(t)

	res := do(t, http.MethodPost, "/deals", map[string]any{
		"make": "Toyota",
	}, cookie)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", res.StatusCode)
	}
}

func TestDealReferrals(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)

	clientID := createTestClient(t, cookie, "555-1005", "deal5@example.com")

	dealRes := do(t, http.MethodPost, "/deals", map[string]any{
		"client_id": clientID,
	}, cookie)
	var deal map[string]any
	decodeJSON(t, dealRes, &deal)
	dealID := deal["id"].(string)

	// Create referral
	refRes := do(t, http.MethodPost, "/deals/"+dealID+"/referrals", map[string]any{
		"first_name": "Refer",
		"last_name":  "Friend",
		"phone":      "555-9999",
		"amount":     500.0,
	}, cookie)
	if refRes.StatusCode != http.StatusCreated {
		t.Fatalf("create referral: got %d, want 201", refRes.StatusCode)
	}
	var ref map[string]any
	decodeJSON(t, refRes, &ref)
	if ref["first_name"] != "Refer" {
		t.Fatalf("first_name = %v, want Refer", ref["first_name"])
	}

	// List referrals
	listRes := do(t, http.MethodGet, "/deals/"+dealID+"/referrals", nil, cookie)
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("list referrals: got %d, want 200", listRes.StatusCode)
	}
	var refs []any
	decodeJSON(t, listRes, &refs)
	if len(refs) != 1 {
		t.Fatalf("expected 1 referral, got %d", len(refs))
	}
}
