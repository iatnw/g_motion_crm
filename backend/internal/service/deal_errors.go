package service

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// DealReferenceError identifies a related record that no longer exists.
// Keeping it independent of HTTP lets API and intake callers share the rule.
type DealReferenceError struct {
	Field string
	cause error
}

func (e *DealReferenceError) Error() string {
	return fmt.Sprintf("%s does not reference an existing record", e.Field)
}

func (e *DealReferenceError) Unwrap() error { return e.cause }

func dealWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return err
	}
	var field string
	switch pgErr.ConstraintName {
	case "deals_client_id_fkey":
		field = "client_id"
	case "deals_broker_id_fkey":
		field = "broker_id"
	case "deals_dealer_id_fkey":
		field = "dealer_id"
	default:
		return err
	}
	return &DealReferenceError{Field: field, cause: err}
}
