package handler

import (
	"testing"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func TestNullPaymentStatusRejectsInvalidValue(t *testing.T) {
	status := "not_paid"

	_, err := nullPaymentStatus(&status)
	if err == nil {
		t.Fatal("expected invalid payment status to be rejected")
	}
}

func TestNullPaymentStatusAcceptsKnownValues(t *testing.T) {
	status := string(db.PaymentStatusPending)

	got, err := nullPaymentStatus(&status)
	if err != nil {
		t.Fatalf("expected payment status to be accepted: %v", err)
	}
	if !got.Valid || got.PaymentStatus != db.PaymentStatusPending {
		t.Fatalf("unexpected payment status: %+v", got)
	}
}

func TestNullReferralStatusRejectsInvalidValue(t *testing.T) {
	status := "not_paid"

	_, err := nullReferralStatus(&status)
	if err == nil {
		t.Fatal("expected invalid referral status to be rejected")
	}
}
