package service

import (
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toNullUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func fromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func toNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	rat := new(big.Rat).SetFloat64(*f)
	num := rat.Num()
	den := rat.Denom()
	exp := int32(0)
	d := new(big.Int).Set(den)
	ten := big.NewInt(10)
	for d.Cmp(big.NewInt(1)) > 0 {
		d.Div(d, ten)
		exp++
	}
	intVal := new(big.Int).Mul(num, new(big.Int).Div(den, new(big.Int).Exp(ten, big.NewInt(int64(exp)), nil)))
	return pgtype.Numeric{Int: intVal, Exp: -exp, Valid: true}
}
