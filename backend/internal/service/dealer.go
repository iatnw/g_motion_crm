package service

import (
	"context"

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
	Name               string
	ContactPerson      *string
	Phone              *string
	Address            *string
	Notes              *string
	ContactPersonEmail *string
}

func (s *DealerService) Create(ctx context.Context, in CreateDealerInput) (db.Dealer, error) {
	return s.q.CreateDealer(ctx, db.CreateDealerParams{
		Name:               in.Name,
		ContactPerson:      toText(in.ContactPerson),
		Phone:              toText(in.Phone),
		Address:            toText(in.Address),
		Notes:              toText(in.Notes),
		ContactPersonEmail: toText(in.ContactPersonEmail),
	})
}

func (s *DealerService) Get(ctx context.Context, id uuid.UUID) (db.Dealer, error) {
	return s.q.GetDealer(ctx, toUUID(id))
}

func (s *DealerService) List(ctx context.Context, limit, offset int32) ([]db.Dealer, error) {
	return s.q.ListDealers(ctx, db.ListDealersParams{Limit: limit, Offset: offset})
}

type UpdateDealerInput struct {
	ID                 uuid.UUID
	Name               string
	ContactPerson      *string
	Phone              *string
	Address            *string
	Notes              *string
	ContactPersonEmail *string
}

func (s *DealerService) Update(ctx context.Context, in UpdateDealerInput) (db.Dealer, error) {
	return s.q.UpdateDealer(ctx, db.UpdateDealerParams{
		ID:                 toUUID(in.ID),
		Name:               in.Name,
		ContactPerson:      toText(in.ContactPerson),
		Phone:              toText(in.Phone),
		Address:            toText(in.Address),
		Notes:              toText(in.Notes),
		ContactPersonEmail: toText(in.ContactPersonEmail),
	})
}

func (s *DealerService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.q.DeleteDealer(ctx, toUUID(id))
	return err
}

func (s *DealerService) GetStats(ctx context.Context, id uuid.UUID) (db.DealerStat, error) {
	return s.q.GetDealerStats(ctx, toUUID(id))
}
