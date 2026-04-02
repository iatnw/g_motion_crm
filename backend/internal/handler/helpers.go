package handler

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

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

func nullDownPaymentMethod(s *string) db.NullDownPaymentMethod {
	if s == nil {
		return db.NullDownPaymentMethod{}
	}
	return db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethod(*s), Valid: true}
}

func nullPaymentStatus(s *string) db.NullPaymentStatus {
	if s == nil {
		return db.NullPaymentStatus{}
	}
	return db.NullPaymentStatus{PaymentStatus: db.PaymentStatus(*s), Valid: true}
}

func nullReferralStatus(s *string) db.NullReferralStatus {
	if s == nil {
		return db.NullReferralStatus{}
	}
	return db.NullReferralStatus{ReferralStatus: db.ReferralStatus(*s), Valid: true}
}

func nullClientsSource(s *string) db.NullClientsSource {
	if s == nil {
		return db.NullClientsSource{}
	}
	return db.NullClientsSource{ClientsSource: db.ClientsSource(*s), Valid: true}
}

func nullEmploymentStatus(s *string) db.NullEmploymentStatus {
	if s == nil {
		return db.NullEmploymentStatus{}
	}
	return db.NullEmploymentStatus{EmploymentStatus: db.EmploymentStatus(*s), Valid: true}
}
