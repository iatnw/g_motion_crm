package handler

import (
	"net/http/httptest"
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

func TestPendingPaymentAgeInvalidDateReturnsNil(t *testing.T) {
	days, bucket := pendingPaymentAge(pgtype.Date{Valid: false})
	if days != nil || bucket != nil {
		t.Fatalf("invalid date: got days=%v bucket=%v, want both nil", days, bucket)
	}
}

func TestPendingPaymentAgeZeroDays(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	days, bucket := pendingPaymentAge(pgtype.Date{Time: now, Valid: true})
	if days == nil || *days != 0 {
		t.Fatalf("days = %v, want 0", days)
	}
	if bucket == nil || *bucket != "green" {
		t.Fatalf("bucket = %v, want green", bucket)
	}
}

func TestNullUUIDValidInput(t *testing.T) {
	s := "550e8400-e29b-41d4-a716-446655440000"
	id, err := nullUUID(&s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == nil {
		t.Fatal("expected non-nil UUID")
	}
}

func TestNullUUIDInvalidInput(t *testing.T) {
	s := "not-a-uuid"
	_, err := nullUUID(&s)
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestNullUUIDNilInput(t *testing.T) {
	id, err := nullUUID(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != nil {
		t.Fatal("expected nil UUID for nil input")
	}
}

func TestNullTimeValidDate(t *testing.T) {
	s := "2026-01-15"
	got, err := nullTime(&s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Format("2006-01-02") != "2026-01-15" {
		t.Fatalf("got %v, want 2026-01-15", got)
	}
}

func TestNullTimeInvalidFormat(t *testing.T) {
	s := "15/01/2026"
	_, err := nullTime(&s)
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestNullTimeNilInput(t *testing.T) {
	got, err := nullTime(nil)
	if err != nil || got != nil {
		t.Fatalf("nil input: got %v err %v, want both nil", got, err)
	}
}

func TestNullDealStatusAcceptsValid(t *testing.T) {
	s := string(db.DealStatusPaid)
	got, err := nullDealStatus(&s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid || got.DealStatus != db.DealStatusPaid {
		t.Fatalf("unexpected deal status: %+v", got)
	}
}

func TestNullDealStatusRejectsInvalid(t *testing.T) {
	s := "invalid_status"
	_, err := nullDealStatus(&s)
	if err == nil {
		t.Fatal("expected error for invalid deal status")
	}
}

func TestNullDownPaymentMethodAcceptsValid(t *testing.T) {
	s := string(db.DownPaymentMethodCash)
	got, err := nullDownPaymentMethod(&s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid || got.DownPaymentMethod != db.DownPaymentMethodCash {
		t.Fatalf("unexpected down payment method: %+v", got)
	}
}

func TestNullDownPaymentMethodRejectsInvalid(t *testing.T) {
	s := "bitcoin"
	_, err := nullDownPaymentMethod(&s)
	if err == nil {
		t.Fatal("expected error for invalid down payment method")
	}
}

func TestNullBrokerPaymentMethodAcceptsValid(t *testing.T) {
	s := string(db.BrokerPaymentMethodCash)
	got, err := nullBrokerPaymentMethod(&s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid || got.BrokerPaymentMethod != db.BrokerPaymentMethodCash {
		t.Fatalf("unexpected broker payment method: %+v", got)
	}
}

func TestNullBrokerPaymentMethodRejectsInvalid(t *testing.T) {
	s := "venmo"
	_, err := nullBrokerPaymentMethod(&s)
	if err == nil {
		t.Fatal("expected error for invalid broker payment method")
	}
}

func TestPaginateExactlyAtLimit(t *testing.T) {
	items := []int{1, 2, 3}
	got, hasMore := paginate(items, 3)
	if len(got) != 3 || hasMore {
		t.Fatalf("got len=%d hasMore=%v, want len=3 hasMore=false", len(got), hasMore)
	}
}

func TestPaginateOverLimit(t *testing.T) {
	items := []int{1, 2, 3, 4}
	got, hasMore := paginate(items, 3)
	if len(got) != 3 || !hasMore {
		t.Fatalf("got len=%d hasMore=%v, want len=3 hasMore=true", len(got), hasMore)
	}
}

func TestParsePaginationDefaults(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	limit, offset := parsePagination(r)
	if limit != defaultLimit || offset != 0 {
		t.Fatalf("got limit=%d offset=%d, want %d/0", limit, offset, defaultLimit)
	}
}

func TestParsePaginationCustomValues(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=10&offset=20", nil)
	limit, offset := parsePagination(r)
	if limit != 10 || offset != 20 {
		t.Fatalf("got limit=%d offset=%d, want 10/20", limit, offset)
	}
}

func TestParsePaginationCapsAtMax(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=9999", nil)
	limit, _ := parsePagination(r)
	if limit != maxLimit {
		t.Fatalf("got limit=%d, want %d (maxLimit)", limit, maxLimit)
	}
}
