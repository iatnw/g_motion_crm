package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type ReferralService struct {
	q *db.Queries
}

func NewReferral(q *db.Queries) *ReferralService {
	return &ReferralService{q: q}
}

type CreateReferralInput struct {
	DealID    *uuid.UUID
	FirstName string
	LastName  string
	Phone     *string
	Amount    *float64
	Notes     *string
}

func (s *ReferralService) Create(ctx context.Context, in CreateReferralInput) (db.Referral, error) {
	return s.q.CreateReferral(ctx, db.CreateReferralParams{
		DealID:    toNullUUID(in.DealID),
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     toText(in.Phone),
		Amount:    toNumeric(in.Amount),
		Status:    db.ReferralStatusNone,
		Notes:     toText(in.Notes),
	})
}

func (s *ReferralService) Get(ctx context.Context, id uuid.UUID) (db.Referral, error) {
	return s.q.GetReferral(ctx, toUUID(id))
}

func (s *ReferralService) List(ctx context.Context, id uuid.UUID) ([]db.Referral, error) {
	return s.q.ListReferralsByDeal(ctx, toUUID(id))
}

func (s *ReferralService) ListAll(ctx context.Context) ([]db.ListAllReferralsRow, error) {
	return s.q.ListAllReferrals(ctx)
}

type UpdateReferralInput struct {
	ID        uuid.UUID
	DealID    *uuid.UUID
	FirstName string
	LastName  string
	Phone     *string
	Amount    *float64
	Notes     *string
}

func (s *ReferralService) Update(ctx context.Context, in UpdateReferralInput) (db.Referral, error) {
	return s.q.UpdateReferral(ctx, db.UpdateReferralParams{
		ID:        toUUID(in.ID),
		DealID:    toNullUUID(in.DealID),
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     toText(in.Phone),
		Amount:    toNumeric(in.Amount),
		Notes:     toText(in.Notes),
	})
}

func (s *ReferralService) UpdateStatus(ctx context.Context, id uuid.UUID, status db.ReferralStatus) (db.Referral, error) {
	return s.q.UpdateReferralStatus(ctx, db.UpdateReferralStatusParams{
		ID:     toUUID(id),
		Status: status,
	})
}

func (s *ReferralService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.q.DeleteReferral(ctx, toUUID(id))
	return err
}
