package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func (s *DealService) WithTx(tx pgx.Tx) *DealService {
	return &DealService{q: s.q.WithTx(tx)}
}

type DealService struct {
	q *db.Queries
}

func NewDeal(q *db.Queries) *DealService {
	return &DealService{q: q}
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
	ReferralAmount         *float64
	YourNet                *float64
	BrokerFeeStatus        db.NullPaymentStatus
	ReferralStatus         db.NullReferralStatus
}

func (s *DealService) Create(ctx context.Context, in CreateDealInput) (db.Deal, error) {
	return s.q.CreateDeal(ctx, db.CreateDealParams{
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
		ReferralAmount:         toNumeric(in.ReferralAmount),
		YourNet:                toNumeric(in.YourNet),
		BrokerFeeStatus:        in.BrokerFeeStatus,
		ReferralStatus:         in.ReferralStatus,
	})
}

func (s *DealService) Get(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.GetDeal(ctx, toUUID(id))
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
	ReferralAmount         *float64
	YourNet                *float64
	BrokerFeeStatus        db.NullPaymentStatus
	ReferralStatus         db.NullReferralStatus
}

func (s *DealService) Update(ctx context.Context, in UpdateDealInput) (db.Deal, error) {
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
		ReferralAmount:         toNumeric(in.ReferralAmount),
		YourNet:                toNumeric(in.YourNet),
		BrokerFeeStatus:        in.BrokerFeeStatus,
		ReferralStatus:         in.ReferralStatus,
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

func (s *DealService) UpdateReferralStatus(ctx context.Context, id uuid.UUID, status db.ReferralStatus) (db.Deal, error) {
	return s.q.UpdateDealReferralStatus(ctx, db.UpdateDealReferralStatusParams{
		ID:             toUUID(id),
		ReferralStatus: db.NullReferralStatus{ReferralStatus: status, Valid: true},
	})
}
