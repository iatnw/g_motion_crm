package service

import (
	"context"
	"database/sql"

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
	Phone     sql.NullString
	Email     sql.NullString
	Notes     sql.NullString
}

func (s *BrokerService) Create(ctx context.Context, in CreateBrokerInput) (db.Broker, error) {
	return s.q.CreateBroker(ctx, db.CreateBrokerParams{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     in.Phone,
		Email:     in.Email,
		Notes:     in.Notes,
	})
}

func (s *BrokerService) Get(ctx context.Context, id uuid.UUID) (db.Broker, error) {
	return s.q.GetBroker(ctx, id)
}

func (s *BrokerService) List(ctx context.Context) ([]db.Broker, error) {
	return s.q.ListBrokers(ctx)
}

type UpdateBrokerInput struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Phone     sql.NullString
	Email     sql.NullString
	Notes     sql.NullString
}

func (s *BrokerService) Update(ctx context.Context, in UpdateBrokerInput) (db.Broker, error) {
	return s.q.UpdateBroker(ctx, db.UpdateBrokerParams{
		ID:        in.ID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Phone:     in.Phone,
		Email:     in.Email,
		Notes:     in.Notes,
	})
}
