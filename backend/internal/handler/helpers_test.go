package handler

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func TestPendingPaymentAgeBuckets(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	tests := []struct {
		name   string
		days   int
		bucket string
	}{
		{name: "green through 31", days: 31, bucket: "green"},
		{name: "yellow starts at 32", days: 32, bucket: "yellow"},
		{name: "red after 40", days: 41, bucket: "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, bucket := pendingPaymentAge(pgtype.Date{
				Time:  now.AddDate(0, 0, -tt.days),
				Valid: true,
			})
			if bucket == nil || *bucket != tt.bucket {
				t.Fatalf("bucket = %v, want %s", bucket, tt.bucket)
			}
		})
	}
}
