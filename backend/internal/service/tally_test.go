package service

import (
	"context"
	"testing"
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

func TestSplitVehicleKeepsYearWithModel(t *testing.T) {
	make, model := splitVehicle("2024 Toyota Camry SE")

	if make != "Toyota" {
		t.Fatalf("make = %q, want Toyota", make)
	}
	if model != "2024 Camry SE" {
		t.Fatalf("model = %q, want 2024 Camry SE", model)
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
