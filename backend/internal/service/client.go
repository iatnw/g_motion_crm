package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func (s *ClientService) WithTx(tx pgx.Tx) *ClientService {
	return &ClientService{q: s.q.WithTx(tx), encryptionKey: s.encryptionKey}
}

type ClientService struct {
	q             *db.Queries
	encryptionKey []byte
}

func NewClient(q *db.Queries, encryptionKey []byte) *ClientService {
	return &ClientService{q: q, encryptionKey: encryptionKey}
}

type CreateClientInput struct {
	FirstName             string
	LastName              string
	Phone                 string
	Email                 string
	Source                db.NullClientsSource
	Status                db.ClientStatus
	Notes                 *string
	DateOfBirth           *time.Time
	Address               *string
	City                  *string
	State                 *string
	Zip                   *string
	Ssn                   *string
	HousingStatus         *string
	HousingDuration       *string
	HousingMonthlyPayment *float64
	EmployerName          *string
	EmployerAddress       *string
	EmployerCity          *string
	EmployerState         *string
	EmployerZip           *string
	Occupation            *string
	EmployerPhone         *string
	EmploymentStatus      db.NullEmploymentStatus
	EmploymentDuration    *string
	GrossMonthlyIncome    *float64
	OtherMonthlyIncome    *float64
	OtherIncomeSource     *string
}

func (s *ClientService) Create(ctx context.Context, in CreateClientInput) (db.Client, error) {
	if in.Ssn != nil {
		enc, err := encrypt(s.encryptionKey, *in.Ssn)
		if err != nil {
			return db.Client{}, err
		}
		in.Ssn = &enc
	}

	client, err := s.q.CreateClient(ctx, db.CreateClientParams{
		FirstName:             in.FirstName,
		LastName:              in.LastName,
		Phone:                 in.Phone,
		Email:                 in.Email,
		Source:                in.Source,
		Status:                in.Status,
		Notes:                 toText(in.Notes),
		DateOfBirth:           toDate(in.DateOfBirth),
		Address:               toText(in.Address),
		City:                  toText(in.City),
		State:                 toText(in.State),
		Zip:                   toText(in.Zip),
		Ssn:                   toText(in.Ssn),
		HousingStatus:         toText(in.HousingStatus),
		HousingDuration:       toText(in.HousingDuration),
		HousingMonthlyPayment: toNumeric(in.HousingMonthlyPayment),
		EmployerName:          toText(in.EmployerName),
		EmployerAddress:       toText(in.EmployerAddress),
		EmployerCity:          toText(in.EmployerCity),
		EmployerState:         toText(in.EmployerState),
		EmployerZip:           toText(in.EmployerZip),
		Occupation:            toText(in.Occupation),
		EmployerPhone:         toText(in.EmployerPhone),
		EmploymentStatus:      in.EmploymentStatus,
		EmploymentDuration:    toText(in.EmploymentDuration),
		GrossMonthlyIncome:    toNumeric(in.GrossMonthlyIncome),
		OtherMonthlyIncome:    toNumeric(in.OtherMonthlyIncome),
		OtherIncomeSource:     toText(in.OtherIncomeSource),
	})
	if err != nil {
		return db.Client{}, err
	}

	return s.decryptSSN(client)
}

func (s *ClientService) Get(ctx context.Context, id uuid.UUID) (db.Client, error) {
	client, err := s.q.GetClient(ctx, toUUID(id))
	if err != nil {
		return db.Client{}, err
	}
	return s.decryptSSN(client)
}

func (s *ClientService) List(ctx context.Context, limit, offset int32) ([]db.Client, error) {
	clients, err := s.q.ListClients(ctx, db.ListClientsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	for i := range clients {
		clients[i], err = s.decryptSSN(clients[i])
		if err != nil {
			return nil, err
		}
	}

	return clients, nil
}

type UpdateClientInput struct {
	ID                    uuid.UUID
	FirstName             string
	LastName              string
	Phone                 string
	Email                 string
	Source                db.NullClientsSource
	Status                db.ClientStatus
	Notes                 *string
	DateOfBirth           *time.Time
	Address               *string
	City                  *string
	State                 *string
	Zip                   *string
	Ssn                   *string
	HousingStatus         *string
	HousingDuration       *string
	HousingMonthlyPayment *float64
	EmployerName          *string
	EmployerAddress       *string
	EmployerCity          *string
	EmployerState         *string
	EmployerZip           *string
	Occupation            *string
	EmployerPhone         *string
	EmploymentStatus      db.NullEmploymentStatus
	EmploymentDuration    *string
	GrossMonthlyIncome    *float64
	OtherMonthlyIncome    *float64
	OtherIncomeSource     *string
}

func (s *ClientService) Update(ctx context.Context, in UpdateClientInput) (db.Client, error) {
	if in.Ssn != nil {
		enc, err := encrypt(s.encryptionKey, *in.Ssn)
		if err != nil {
			return db.Client{}, err
		}
		in.Ssn = &enc
	}

	client, err := s.q.UpdateClient(ctx, db.UpdateClientParams{
		ID:                    toUUID(in.ID),
		FirstName:             in.FirstName,
		LastName:              in.LastName,
		Phone:                 in.Phone,
		Email:                 in.Email,
		Source:                in.Source,
		Status:                in.Status,
		Notes:                 toText(in.Notes),
		DateOfBirth:           toDate(in.DateOfBirth),
		Address:               toText(in.Address),
		City:                  toText(in.City),
		State:                 toText(in.State),
		Zip:                   toText(in.Zip),
		Ssn:                   toText(in.Ssn),
		HousingStatus:         toText(in.HousingStatus),
		HousingDuration:       toText(in.HousingDuration),
		HousingMonthlyPayment: toNumeric(in.HousingMonthlyPayment),
		EmployerName:          toText(in.EmployerName),
		EmployerAddress:       toText(in.EmployerAddress),
		EmployerCity:          toText(in.EmployerCity),
		EmployerState:         toText(in.EmployerState),
		EmployerZip:           toText(in.EmployerZip),
		Occupation:            toText(in.Occupation),
		EmployerPhone:         toText(in.EmployerPhone),
		EmploymentStatus:      in.EmploymentStatus,
		EmploymentDuration:    toText(in.EmploymentDuration),
		GrossMonthlyIncome:    toNumeric(in.GrossMonthlyIncome),
		OtherMonthlyIncome:    toNumeric(in.OtherMonthlyIncome),
		OtherIncomeSource:     toText(in.OtherIncomeSource),
	})
	if err != nil {
		return db.Client{}, err
	}

	return s.decryptSSN(client)
}

func (s *ClientService) decryptSSN(client db.Client) (db.Client, error) {
	if !client.Ssn.Valid {
		return client, nil
	}
	plain, err := decrypt(s.encryptionKey, client.Ssn.String)
	if err != nil {
		return db.Client{}, err
	}
	client.Ssn.String = plain
	return client, nil
}
