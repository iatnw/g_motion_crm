package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDealWriteError(t *testing.T) {
	for _, field := range []string{"client_id", "broker_id", "dealer_id"} {
		t.Run(field, func(t *testing.T) {
			cause := &pgconn.PgError{Code: "23503", ConstraintName: "deals_" + field + "_fkey"}
			err := dealWriteError(fmt.Errorf("insert deal: %w", cause))
			var ref *DealReferenceError
			if !errors.As(err, &ref) || ref.Field != field {
				t.Fatalf("unexpected error: %v", err)
			}
			if !errors.Is(err, cause) {
				t.Fatal("database cause was lost")
			}
		})
	}
	for _, original := range []error{nil, errors.New("connection unavailable"), &pgconn.PgError{Code: "23505", ConstraintName: "deals_broker_id_fkey"}, &pgconn.PgError{Code: "23503", ConstraintName: "other_fkey"}} {
		if got := dealWriteError(original); got != original {
			t.Fatalf("unrelated error changed: %v", got)
		}
	}
}
