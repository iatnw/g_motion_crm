package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type DealerService struct {
	q *db.Queries
}

func NewDealer(q *db.Queries) *DealerService {
	return &DealerService{q: q}
}

type CreateDealerInput struct {
	Name          string
	ContactPerson sql.NullString
	Phone         sql.NullString
	Address       sql.NullString
	Notes         sql.NullString
}

func (s *DealerService) Create(ctx context.Context, in CreateDealerInput) (db.Dealer, error) {
	return s.q.CreateDealer(ctx, db.CreateDealerParams{
		Name:          in.Name,
		ContactPerson: in.ContactPerson,
		Phone:         in.Phone,
		Address:       in.Address,
		Notes:         in.Notes,
	})
}

func (s *DealerService) Get(ctx context.Context, id uuid.UUID) (db.Dealer, error) {
	return s.q.GetDealer(ctx, id)
}

func (s *DealerService) List(ctx context.Context, limit, offset int32) ([]db.Dealer, error) {
	return s.q.ListDealers(ctx, db.ListDealersParams{Limit: limit, Offset: offset})
}

type UpdateDealerInput struct {
	ID            uuid.UUID
	Name          string
	ContactPerson sql.NullString
	Phone         sql.NullString
	Address       sql.NullString
	Notes         sql.NullString
}

func (s *DealerService) Update(ctx context.Context, in UpdateDealerInput) (db.Dealer, error) {
	return s.q.UpdateDealer(ctx, db.UpdateDealerParams{
		ID:            in.ID,
		Name:          in.Name,
		ContactPerson: in.ContactPerson,
		Phone:         in.Phone,
		Address:       in.Address,
		Notes:         in.Notes,
	})
}

func (s *DealerService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.q.DeleteDealer(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
