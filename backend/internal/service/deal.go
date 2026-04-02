package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type DealService struct {
	q *db.Queries
}

func NewDeal(q *db.Queries) *DealService {
	return &DealService{q: q}
}

type CreateDealInput struct {
	ClientID               uuid.UUID
	BrokerID               uuid.NullUUID
	DealerID               uuid.NullUUID
	DealDate               sql.NullTime
	DealNumber             sql.NullString
	Vin                    sql.NullString
	Make                   sql.NullString
	Model                  sql.NullString
	TrimLevel              sql.NullString
	DownPayment            sql.NullString
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         sql.NullString
	BrokerFee              sql.NullString
	BrokerPercent          sql.NullString
	BrokerCommission       sql.NullString
	BrokerCommissionStatus db.NullPaymentStatus
	ReferralAmount         sql.NullString
	YourNet                sql.NullString
	BrokerFeeStatus        db.NullPaymentStatus
	ReferralStatus         db.NullReferralStatus
}

func (s *DealService) Create(ctx context.Context, in CreateDealInput) (db.Deal, error) {
	return s.q.CreateDeal(ctx, db.CreateDealParams{
		ClientID:               in.ClientID,
		BrokerID:               in.BrokerID,
		DealerID:               in.DealerID,
		DealDate:               in.DealDate,
		DealNumber:             in.DealNumber,
		Vin:                    in.Vin,
		Make:                   in.Make,
		Model:                  in.Model,
		TrimLevel:              in.TrimLevel,
		DownPayment:            in.DownPayment,
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         in.MonthlyPayment,
		BrokerFee:              in.BrokerFee,
		BrokerPercent:          in.BrokerPercent,
		BrokerCommission:       in.BrokerCommission,
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		ReferralAmount:         in.ReferralAmount,
		YourNet:                in.YourNet,
		BrokerFeeStatus:        in.BrokerFeeStatus,
		ReferralStatus:         in.ReferralStatus,
	})
}

func (s *DealService) Get(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.GetDeal(ctx, id)
}

func (s *DealService) List(ctx context.Context) ([]db.Deal, error) {
	return s.q.ListDeals(ctx)
}

func (s *DealService) ListByClient(ctx context.Context, id uuid.UUID) ([]db.Deal, error) {
	return s.q.ListDealsByClient(ctx, id)
}

type UpdateDealInput struct {
	ID                     uuid.UUID
	ClientID               uuid.UUID
	BrokerID               uuid.NullUUID
	DealerID               uuid.NullUUID
	DealDate               sql.NullTime
	DealNumber             sql.NullString
	Vin                    sql.NullString
	Make                   sql.NullString
	Model                  sql.NullString
	TrimLevel              sql.NullString
	DownPayment            sql.NullString
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         sql.NullString
	BrokerFee              sql.NullString
	BrokerPercent          sql.NullString
	BrokerCommission       sql.NullString
	BrokerCommissionStatus db.NullPaymentStatus
	ReferralAmount         sql.NullString
	YourNet                sql.NullString
	BrokerFeeStatus        db.NullPaymentStatus
	ReferralStatus         db.NullReferralStatus
}

func (s *DealService) Update(ctx context.Context, in UpdateDealInput) (db.Deal, error) {
	return s.q.UpdateDeal(ctx, db.UpdateDealParams{
		ID:                     in.ID,
		ClientID:               in.ClientID,
		BrokerID:               in.BrokerID,
		DealerID:               in.DealerID,
		DealDate:               in.DealDate,
		DealNumber:             in.DealNumber,
		Vin:                    in.Vin,
		Make:                   in.Make,
		Model:                  in.Model,
		TrimLevel:              in.TrimLevel,
		DownPayment:            in.DownPayment,
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         in.MonthlyPayment,
		BrokerFee:              in.BrokerFee,
		BrokerPercent:          in.BrokerPercent,
		BrokerCommission:       in.BrokerCommission,
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		ReferralAmount:         in.ReferralAmount,
		YourNet:                in.YourNet,
		BrokerFeeStatus:        in.BrokerFeeStatus,
		ReferralStatus:         in.ReferralStatus,
	})
}

func (s *DealService) UpdateBrokerFeeStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus) (db.Deal, error) {
	return s.q.UpdateDealBrokerFeeStatus(ctx, db.UpdateDealBrokerFeeStatusParams{
		ID:              id,
		BrokerFeeStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
	})
}

func (s *DealService) UpdateBrokerCommissionStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus) (db.Deal, error) {
	return s.q.UpdateDealBrokerCommissionStatus(ctx, db.UpdateDealBrokerCommissionStatusParams{
		ID:                     id,
		BrokerCommissionStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
	})
}

func (s *DealService) UpdateReferralStatus(ctx context.Context, id uuid.UUID, status db.ReferralStatus) (db.Deal, error) {
	return s.q.UpdateDealReferralStatus(ctx, db.UpdateDealReferralStatusParams{
		ID:             id,
		ReferralStatus: db.NullReferralStatus{ReferralStatus: status, Valid: true},
	})
}
