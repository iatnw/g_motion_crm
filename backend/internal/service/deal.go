package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

var ErrBrokerCommissionMismatch = errors.New("broker_commission does not match broker_percent and broker_fee")

func (s *DealService) WithTx(tx pgx.Tx) *DealService {
	return &DealService{q: s.q.WithTx(tx), pool: s.pool}
}

type DealService struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewDeal(q *db.Queries, pool *pgxpool.Pool) *DealService {
	return &DealService{q: q, pool: pool}
}

func (s *DealService) PreviewNextDealNumber(ctx context.Context) (string, error) {
	var seq int64
	err := s.pool.QueryRow(ctx,
		`SELECT CASE WHEN is_called THEN last_value + 1 ELSE last_value END FROM deal_number_seq`,
	).Scan(&seq)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("GM-%d-%04d", time.Now().Year(), seq), nil
}

type CreateDealInput struct {
	ClientID               uuid.UUID
	BrokerID               *uuid.UUID
	DealerID               *uuid.UUID
	DealDate               *time.Time
	DealNumber             *string
	Vin                    *string
	Make                   *string
	Model                  *string
	TrimLevel              *string
	DownPayment            *float64
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         *float64
	BrokerFee              *float64
	BrokerPercent          *float64
	BrokerCommission       *float64
	BrokerCommissionStatus db.NullPaymentStatus
	YourNet                *float64
	BrokerFeeStatus        db.NullPaymentStatus
	DealStatus             db.NullDealStatus
	CollectAmount          *float64
	DocFee                 *float64
	BrokerPaymentMethod    db.NullBrokerPaymentMethod
	Notes                  *string
}

func (s *DealService) Create(ctx context.Context, in CreateDealInput) (db.Deal, error) {
	if !in.DealStatus.Valid {
		in.DealStatus = db.NullDealStatus{DealStatus: db.DealStatusPending, Valid: true}
	}
	if !in.BrokerFeeStatus.Valid {
		in.BrokerFeeStatus = db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPending, Valid: true}
	}
	if !in.BrokerCommissionStatus.Valid {
		in.BrokerCommissionStatus = db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPending, Valid: true}
	}
	if in.DealDate == nil {
		now := time.Now()
		in.DealDate = &now
	}

	if err := normalizeBrokerCommission(in.BrokerFee, in.CollectAmount, &in.BrokerPercent, &in.BrokerCommission, &in.YourNet); err != nil {
		return db.Deal{}, err
	}

	if in.DealNumber == nil {
		seq, err := s.q.NextDealNumber(ctx)
		if err != nil {
			return db.Deal{}, err
		}
		year := time.Now().Year()
		num := fmt.Sprintf("GM-%d-%04d", year, seq)
		in.DealNumber = &num
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Deal{}, err
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)
	deal, err := q.CreateDeal(ctx, db.CreateDealParams{
		ClientID:               toUUID(in.ClientID),
		BrokerID:               toNullUUID(in.BrokerID),
		DealerID:               toNullUUID(in.DealerID),
		DealDate:               toDate(in.DealDate),
		DealNumber:             toText(in.DealNumber),
		Vin:                    toText(in.Vin),
		Make:                   toText(in.Make),
		Model:                  toText(in.Model),
		TrimLevel:              toText(in.TrimLevel),
		DownPayment:            toNumeric(in.DownPayment),
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         toNumeric(in.MonthlyPayment),
		BrokerFee:              toNumeric(in.BrokerFee),
		BrokerPercent:          toNumeric(in.BrokerPercent),
		BrokerCommission:       toNumeric(in.BrokerCommission),
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		YourNet:                toNumeric(in.YourNet),
		BrokerFeeStatus:        in.BrokerFeeStatus,
		DealStatus:             in.DealStatus,
		CollectAmount:          toNumeric(in.CollectAmount),
		DocFee:                 toNumeric(in.DocFee),
		BrokerPaymentMethod:    in.BrokerPaymentMethod,
		Notes:                  toText(in.Notes),
	})
	if err != nil {
		return db.Deal{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Deal{}, err
	}
	return deal, nil
}

func (s *DealService) Get(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.GetDeal(ctx, toUUID(id))
}

func (s *DealService) GetShow(ctx context.Context, id uuid.UUID) (db.GetDealShowRow, error) {
	return s.q.GetDealShow(ctx, toUUID(id))
}

func (s *DealService) List(ctx context.Context, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDeals(ctx, db.ListDealsParams{Limit: limit, Offset: offset})
}

func (s *DealService) ListByClient(ctx context.Context, id uuid.UUID, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDealsByClient(ctx, db.ListDealsByClientParams{
		ClientID: toUUID(id),
		Limit:    limit,
		Offset:   offset,
	})
}

type UpdateDealInput struct {
	ID                     uuid.UUID
	ClientID               uuid.UUID
	BrokerID               *uuid.UUID
	DealerID               *uuid.UUID
	DealDate               *time.Time
	DealNumber             *string
	Vin                    *string
	Make                   *string
	Model                  *string
	TrimLevel              *string
	DownPayment            *float64
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         *float64
	BrokerFee              *float64
	BrokerPercent          *float64
	BrokerCommission       *float64
	BrokerCommissionStatus db.NullPaymentStatus
	YourNet                *float64
	BrokerFeeStatus        db.NullPaymentStatus
	DealStatus             db.NullDealStatus
	CollectAmount          *float64
	DocFee                 *float64
	BrokerPaymentMethod    db.NullBrokerPaymentMethod
	Notes                  *string
}

func (s *DealService) Update(ctx context.Context, in UpdateDealInput) (db.Deal, error) {
	if err := normalizeBrokerCommission(in.BrokerFee, in.CollectAmount, &in.BrokerPercent, &in.BrokerCommission, &in.YourNet); err != nil {
		return db.Deal{}, err
	}

	return s.q.UpdateDeal(ctx, db.UpdateDealParams{
		ID:                     toUUID(in.ID),
		ClientID:               toUUID(in.ClientID),
		BrokerID:               toNullUUID(in.BrokerID),
		DealerID:               toNullUUID(in.DealerID),
		DealDate:               toDate(in.DealDate),
		DealNumber:             toText(in.DealNumber),
		Vin:                    toText(in.Vin),
		Make:                   toText(in.Make),
		Model:                  toText(in.Model),
		TrimLevel:              toText(in.TrimLevel),
		DownPayment:            toNumeric(in.DownPayment),
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         toNumeric(in.MonthlyPayment),
		BrokerFee:              toNumeric(in.BrokerFee),
		BrokerPercent:          toNumeric(in.BrokerPercent),
		BrokerCommission:       toNumeric(in.BrokerCommission),
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		YourNet:                toNumeric(in.YourNet),
		BrokerFeeStatus:        in.BrokerFeeStatus,
		DealStatus:             in.DealStatus,
		CollectAmount:          toNumeric(in.CollectAmount),
		DocFee:                 toNumeric(in.DocFee),
		BrokerPaymentMethod:    in.BrokerPaymentMethod,
		Notes:                  toText(in.Notes),
	})
}

func (s *DealService) UpdateBrokerFeeStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus) (db.Deal, error) {
	return s.q.UpdateDealBrokerFeeStatus(ctx, db.UpdateDealBrokerFeeStatusParams{
		ID:              toUUID(id),
		BrokerFeeStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
	})
}

func (s *DealService) UpdateBrokerCommissionStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus) (db.Deal, error) {
	return s.q.UpdateDealBrokerCommissionStatus(ctx, db.UpdateDealBrokerCommissionStatusParams{
		ID:                     toUUID(id),
		BrokerCommissionStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
	})
}

func (s *DealService) MarkPaymentReminderEmailSent(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.MarkDealPaymentReminderEmailSent(ctx, toUUID(id))
}

func normalizeBrokerCommission(brokerFee *float64, collectAmount *float64, brokerPercent **float64, brokerCommission **float64, yourNet **float64) error {
	fee := 0.0
	if brokerFee != nil {
		fee = *brokerFee
	}
	collect := 0.0
	if collectAmount != nil {
		collect = *collectAmount
	}

	base := fee
	if base == 0 {
		base = collect
	}

	hasPercent := brokerPercent != nil && *brokerPercent != nil
	hasCommission := brokerCommission != nil && *brokerCommission != nil

	if hasPercent && !hasCommission {
		if base > 0 {
			commission := base * (**brokerPercent) / 100
			*brokerCommission = &commission
		}
	} else if hasCommission && !hasPercent {
		if base > 0 {
			percent := (**brokerCommission) / base * 100
			*brokerPercent = &percent
		}
	} else if hasPercent && hasCommission {
		if base > 0 {
			expectedCommission := base * (**brokerPercent) / 100
			if math.Abs(expectedCommission-(**brokerCommission)) > 0.01 {
				return ErrBrokerCommissionMismatch
			}
		}
	}

	if yourNet != nil && *yourNet == nil && hasCommission {
		net := base - (**brokerCommission)
		*yourNet = &net
	}

	return nil
}
