package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type BrokerService struct {
	q *db.Queries
}

func NewBroker(q *db.Queries) *BrokerService {
	return &BrokerService{q: q}
}

type CreateBrokerInput struct {
	FirstName string
	LastName  string
	Phone     *string
	Email     *string
	Notes     *string
}

func (s *BrokerService) Create(ctx context.Context, in CreateBrokerInput) (db.Broker, error) {
	return s.q.CreateBroker(ctx, db.CreateBrokerParams{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     toText(in.Phone),
		Email:     toText(in.Email),
		Notes:     toText(in.Notes),
	})
}

func (s *BrokerService) Get(ctx context.Context, id uuid.UUID) (db.Broker, error) {
	return s.q.GetBroker(ctx, toUUID(id))
}

func (s *BrokerService) List(ctx context.Context, limit, offset int32) ([]db.Broker, error) {
	return s.q.ListBrokers(ctx, db.ListBrokersParams{Limit: limit, Offset: offset})
}

type UpdateBrokerInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Phone     *string
	Email     *string
	Notes     *string
}

func (s *BrokerService) Update(ctx context.Context, in UpdateBrokerInput) (db.Broker, error) {
	return s.q.UpdateBroker(ctx, db.UpdateBrokerParams{
		ID:        toUUID(in.ID),
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     toText(in.Phone),
		Email:     toText(in.Email),
		Notes:     toText(in.Notes),
	})
}

func (s *BrokerService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.q.DeleteBroker(ctx, toUUID(id))
	return err
}

func (s *BrokerService) GetStats(ctx context.Context, id uuid.UUID) (db.BrokerStat, error) {
	return s.q.GetBrokerStats(ctx, toUUID(id))
}
