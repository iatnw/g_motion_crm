package handler

import (
	"net/http/httptest"
	"testing"
)

func TestOptionalQueryDate(t *testing.T) {
	r := httptest.NewRequest("GET", "/dashboard/summary?from=2026-04-01", nil)

	got, err := optionalQueryDate(r, "from")
	if err != nil {
		t.Fatalf("optionalQueryDate returned error: %v", err)
	}
	if got == nil || got.Format(dateLayout) != "2026-04-01" {
		t.Fatalf("date = %v, want 2026-04-01", got)
	}
}

func TestOptionalQueryDateRejectsInvalidDate(t *testing.T) {
	r := httptest.NewRequest("GET", "/dashboard/summary?from=04-01-2026", nil)

	_, err := optionalQueryDate(r, "from")
	if err == nil {
		t.Fatal("expected invalid date to be rejected")
	}
}
