package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

func parsePagination(r *http.Request) (limit, offset int32) {
	limit = defaultLimit
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return limit, offset
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullUUID(s *string) (uuid.NullUUID, error) {
	if s == nil {
		return uuid.NullUUID{}, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

const dateLayout = "2006-01-02"

func nullTime(s *string) (sql.NullTime, error) {
	if s == nil {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse(dateLayout, *s)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

func nullDownPaymentMethod(s *string) (db.NullDownPaymentMethod, error) {
	if s == nil {
		return db.NullDownPaymentMethod{}, nil
	}
	switch db.DownPaymentMethod(*s) {
	case db.DownPaymentMethodCash, db.DownPaymentMethodCard:
		return db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethod(*s), Valid: true}, nil
	default:
		return db.NullDownPaymentMethod{}, fmt.Errorf("invalid down payment method: %s", *s)
	}

}

func nullPaymentStatus(s *string) (db.NullPaymentStatus, error) {
	if s == nil {
		return db.NullPaymentStatus{}, nil
	}
	switch db.PaymentStatus(*s) {
	case db.PaymentStatusPending, db.PaymentStatusPaid:
		return db.NullPaymentStatus{PaymentStatus: db.PaymentStatus(*s), Valid: true}, nil
	default:
		return db.NullPaymentStatus{}, fmt.Errorf("invalid payment status: %s", *s)
	}
}

func nullReferralStatus(s *string) (db.NullReferralStatus, error) {
	if s == nil {
		return db.NullReferralStatus{}, nil
	}
	switch db.ReferralStatus(*s) {
	case db.ReferralStatusPending, db.ReferralStatusPaid, db.ReferralStatusNone:
		return db.NullReferralStatus{ReferralStatus: db.ReferralStatus(*s), Valid: true}, nil
	default:
		return db.NullReferralStatus{}, fmt.Errorf("invalid referral status: %s", *s)
	}
}

func nullClientsSource(s *string) (db.NullClientsSource, error) {
	if s == nil {
		return db.NullClientsSource{}, nil
	}
	switch db.ClientsSource(*s) {
	case db.ClientsSourceInstagram, db.ClientsSourceRepeatClient, db.ClientsSourceOrganic, db.ClientsSourceReferall, db.ClientsSourceTelegram:
		return db.NullClientsSource{ClientsSource: db.ClientsSource(*s), Valid: true}, nil
	default:
		return db.NullClientsSource{}, fmt.Errorf("invalid clients source: %s", *s)
	}
}

func nullEmploymentStatus(s *string) (db.NullEmploymentStatus, error) {
	if s == nil {
		return db.NullEmploymentStatus{}, nil
	}
	switch db.EmploymentStatus(*s) {
	case db.EmploymentStatusSelfEmployed, db.EmploymentStatusPartTime, db.EmploymentStatusFullTime:
		return db.NullEmploymentStatus{EmploymentStatus: db.EmploymentStatus(*s), Valid: true}, nil
	default:
		return db.NullEmploymentStatus{}, fmt.Errorf("invalid employment status: %s", *s)
	}
}
