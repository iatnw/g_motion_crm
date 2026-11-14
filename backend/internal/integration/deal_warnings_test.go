package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestVINWarningRules(t *testing.T) {
	cases := []struct {
		vin     *string
		warning bool
	}{
		{nil, false}, {strptr(""), false}, {strptr(" \t\n"), false},
		{strptr("1HGCM82633A004352"), false}, {strptr("1M8GDM9AXKP042788"), false},
		{strptr(" 1hg cm82633a004352\n"), false},
		{strptr("1HGCM82633A004353"), true}, {strptr("1HGCM82643A004352"), true},
		{strptr("1HG"), true}, {strptr("1HGCM82633A0043521"), true},
		{strptr("1HGCM82633A00435I"), true}, {strptr("1HGCM82633A00435O"), true},
		{strptr("1HGCM82633A00435Q"), true}, {strptr("1HGCM82633A00435!"), true},
	}
	for _, c := range cases {
		var got bool
		if err := pool.QueryRow(context.Background(), "SELECT vin_has_warning($1)", c.vin).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != c.warning {
			t.Fatalf("VIN %v: warning %v, want %v", c.vin, got, c.warning)
		}
	}
}

func strptr(s string) *string { return &s }

func TestDealWarningsFilter(t *testing.T) {
	cookie := login(t)
	t.Cleanup(truncateData)
	clientID := createTestClient(t, cookie, "555-8200", "warnings@example.com")
	for i, vin := range []string{"1HGCM82633A004352", "1HGCM82633A004353", "BADVIN", ""} {
		res := do(t, http.MethodPost, "/deals", map[string]any{
			"client_id": clientID, "vin": vin, "make": "WarningTest", "notes": fmt.Sprint(i),
		}, cookie)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create: %d", res.StatusCode)
		}
		res.Body.Close()
	}
	type page struct {
		Data    []map[string]any `json:"data"`
		HasMore bool             `json:"has_more"`
		Counts  map[string]int   `json:"counts"`
	}
	var first page
	res := do(t, http.MethodGet, "/deals?warnings=any&vin=present&limit=1", nil, cookie)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list: %d", res.StatusCode)
	}
	decodeJSON(t, res, &first)
	if len(first.Data) != 1 || !first.HasMore || first.Counts["all"] != 2 || first.Counts["pending"] != 2 || first.Counts["new"] != 0 {
		t.Fatalf("unexpected warning page: %+v", first)
	}
	var second page
	decodeJSON(t, do(t, http.MethodGet, "/deals?warnings=any&vin=present&limit=1&offset=1", nil, cookie), &second)
	if len(second.Data) != 1 || second.HasMore || first.Data[0]["id"] == second.Data[0]["id"] {
		t.Fatalf("bad pagination: %+v", second)
	}
	var searched page
	decodeJSON(t, do(t, http.MethodGet, "/deals?warnings=any&search=BADVIN", nil, cookie), &searched)
	if len(searched.Data) != 1 || searched.Counts["all"] != 1 {
		t.Fatalf("bad search: %+v", searched)
	}
	var missing page
	decodeJSON(t, do(t, http.MethodGet, "/deals?warnings=any&vin=missing", nil, cookie), &missing)
	if len(missing.Data) != 0 {
		t.Fatal("blank VIN should not have a warning")
	}
	var unfiltered page
	decodeJSON(t, do(t, http.MethodGet, "/deals", nil, cookie), &unfiltered)
	if len(unfiltered.Data) != 4 {
		t.Fatalf("default filter hid deals: %+v", unfiltered)
	}
	res = do(t, http.MethodGet, "/deals?warnings=unknown", nil, cookie)
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown warning filter: %d", res.StatusCode)
	}
}
