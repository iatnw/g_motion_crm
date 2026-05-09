package service

import (
	"context"
	"testing"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func TestFieldParserUsesDuplicateLabelOrder(t *testing.T) {
	p := newParser([]TallyField{
		{Label: "City", Value: "Los Angeles"},
		{Label: "State", Value: "CA"},
		{Label: "City", Value: "Burbank"},
	})

	if got := p.str("City"); got != "Los Angeles" {
		t.Fatalf("first city = %q, want Los Angeles", got)
	}
	if got := p.str("City"); got != "Burbank" {
		t.Fatalf("second city = %q, want Burbank", got)
	}
}

func TestFieldParserMissingLabelReturnsEmpty(t *testing.T) {
	p := newParser([]TallyField{{Label: "Name", Value: "John"}})
	if got := p.str("Phone"); got != "" {
		t.Fatalf("missing label = %q, want empty", got)
	}
}

func TestSplitVehicleKeepsYearWithModel(t *testing.T) {
	make, model := splitVehicle("2024 Toyota Camry SE")

	if make != "Toyota" {
		t.Fatalf("make = %q, want Toyota", make)
	}
	if model != "2024 Camry SE" {
		t.Fatalf("model = %q, want 2024 Camry SE", model)
	}
}

func TestSplitVehicleNoYear(t *testing.T) {
	make, model := splitVehicle("Honda Civic")
	if make != "Honda" {
		t.Fatalf("make = %q, want Honda", make)
	}
	if model != "Civic" {
		t.Fatalf("model = %q, want Civic", model)
	}
}

func TestSplitVehicleMakeOnly(t *testing.T) {
	make, model := splitVehicle("Toyota")
	if make != "Toyota" {
		t.Fatalf("make = %q, want Toyota", make)
	}
	if model != "" {
		t.Fatalf("model = %q, want empty", model)
	}
}

func TestSplitVehicleEmpty(t *testing.T) {
	make, model := splitVehicle("")
	if make != "" || model != "" {
		t.Fatalf("empty input: make = %q, model = %q, want both empty", make, model)
	}
}

func TestSplitNameTwoParts(t *testing.T) {
	first, last, err := splitName("John Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != "John" || last != "Doe" {
		t.Fatalf("got %q %q, want John Doe", first, last)
	}
}

func TestSplitNameMultipleWords(t *testing.T) {
	first, last, err := splitName("Mary Jane Watson")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != "Mary" || last != "Jane Watson" {
		t.Fatalf("got %q %q, want Mary / Jane Watson", first, last)
	}
}

func TestSplitNameSingleWord(t *testing.T) {
	first, last, err := splitName("Madonna")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != "Madonna" || last != "Madonna" {
		t.Fatalf("single word: got %q %q, want Madonna/Madonna", first, last)
	}
}

func TestSplitNameEmpty(t *testing.T) {
	_, _, err := splitName("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLabelToDocTypeDriverLicense(t *testing.T) {
	if got := labelToDocType("Driver's License"); got != db.DocumentTypeDriverLicense {
		t.Fatalf("got %q, want driver_license", got)
	}
}

func TestLabelToDocTypeInsurance(t *testing.T) {
	if got := labelToDocType("Insurance"); got != db.DocumentTypeInsurance {
		t.Fatalf("got %q, want insurance", got)
	}
}

func TestLabelToDocTypeUnknownFallsBackToOther(t *testing.T) {
	if got := labelToDocType("some unknown field"); got != db.DocumentTypeOther {
		t.Fatalf("got %q, want other", got)
	}
}

func TestClaimTallyEventRequiresIdentifier(t *testing.T) {
	claimed, err := claimTallyEvent(context.Background(), nil, "", " ")
	if err == nil {
		t.Fatal("expected missing identifiers to fail")
	}
	if claimed {
		t.Fatal("missing identifiers must not be claimed")
	}
}
