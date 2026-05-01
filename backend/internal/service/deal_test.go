package service

import (
	"errors"
	"math"
	"testing"
)

func TestNormalizeBrokerCommissionCalculatesCommission(t *testing.T) {
	fee := 1000.0
	percentValue := 25.0
	percent := &percentValue
	var commission *float64
	var net *float64

	err := normalizeBrokerCommission(&fee, nil, &percent, &commission, &net)
	if err != nil {
		t.Fatalf("normalize broker commission: %v", err)
	}
	if commission == nil || math.Abs(*commission-250) > 0.01 {
		t.Fatalf("commission = %v, want 250", commission)
	}
}

func TestNormalizeBrokerCommissionCalculatesPercent(t *testing.T) {
	fee := 1000.0
	var percent *float64
	commissionValue := 250.0
	commission := &commissionValue
	var net *float64

	err := normalizeBrokerCommission(&fee, nil, &percent, &commission, &net)
	if err != nil {
		t.Fatalf("normalize broker commission: %v", err)
	}
	if percent == nil || math.Abs(*percent-25) > 0.01 {
		t.Fatalf("percent = %v, want 25", percent)
	}
}

func TestNormalizeBrokerCommissionRejectsMismatch(t *testing.T) {
	fee := 1000.0
	percentValue := 25.0
	percent := &percentValue
	commissionValue := 300.0
	commission := &commissionValue
	var net *float64

	err := normalizeBrokerCommission(&fee, nil, &percent, &commission, &net)
	if !errors.Is(err, ErrBrokerCommissionMismatch) {
		t.Fatalf("err = %v, want ErrBrokerCommissionMismatch", err)
	}
}
