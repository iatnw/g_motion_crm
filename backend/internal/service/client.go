package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func (s *ClientService) WithTx(tx *sql.Tx) *ClientService {
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
	Notes                 sql.NullString
	DateOfBirth           sql.NullTime
	Address               sql.NullString
	City                  sql.NullString
	State                 sql.NullString
	Zip                   sql.NullString
	Ssn                   sql.NullString
	HousingStatus         sql.NullString
	HousingDuration       sql.NullString
	HousingMonthlyPayment sql.NullString
	EmployerName          sql.NullString
	EmployerAddress       sql.NullString
	EmployerCity          sql.NullString
	EmployerState         sql.NullString
	EmployerZip           sql.NullString
	Occupation            sql.NullString
	EmployerPhone         sql.NullString
	EmploymentStatus      db.NullEmploymentStatus
	EmploymentDuration    sql.NullString
	GrossMonthlyIncome    sql.NullString
	OtherMonthlyIncome    sql.NullString
	OtherIncomeSource     sql.NullString
}

func (s *ClientService) Create(ctx context.Context, in CreateClientInput) (db.Client, error) {
	if in.Ssn.Valid {
		enc, err := encrypt(s.encryptionKey, in.Ssn.String)
		if err != nil {
			return db.Client{}, err
		}
		in.Ssn.String = enc
	}

	client, err := s.q.CreateClient(ctx, db.CreateClientParams{
		FirstName:             in.FirstName,
		LastName:              in.LastName,
		Phone:                 in.Phone,
		Email:                 in.Email,
		Source:                in.Source,
		Status:                in.Status,
		Notes:                 in.Notes,
		DateOfBirth:           in.DateOfBirth,
		Address:               in.Address,
		City:                  in.City,
		State:                 in.State,
		Zip:                   in.Zip,
		Ssn:                   in.Ssn,
		HousingStatus:         in.HousingStatus,
		HousingDuration:       in.HousingDuration,
		HousingMonthlyPayment: in.HousingMonthlyPayment,
		EmployerName:          in.EmployerName,
		EmployerAddress:       in.EmployerAddress,
		EmployerCity:          in.EmployerCity,
		EmployerState:         in.EmployerState,
		EmployerZip:           in.EmployerZip,
		Occupation:            in.Occupation,
		EmployerPhone:         in.EmployerPhone,
		EmploymentStatus:      in.EmploymentStatus,
		EmploymentDuration:    in.EmploymentDuration,
		GrossMonthlyIncome:    in.GrossMonthlyIncome,
		OtherMonthlyIncome:    in.OtherMonthlyIncome,
		OtherIncomeSource:     in.OtherIncomeSource,
	})
	if err != nil {
		return db.Client{}, err
	}

	return s.decryptSSN(client)
}

func (s *ClientService) Get(ctx context.Context, id uuid.UUID) (db.Client, error) {
	client, err := s.q.GetClient(ctx, id)
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
	Notes                 sql.NullString
	DateOfBirth           sql.NullTime
	Address               sql.NullString
	City                  sql.NullString
	State                 sql.NullString
	Zip                   sql.NullString
	Ssn                   sql.NullString
	HousingStatus         sql.NullString
	HousingDuration       sql.NullString
	HousingMonthlyPayment sql.NullString
	EmployerName          sql.NullString
	EmployerAddress       sql.NullString
	EmployerCity          sql.NullString
	EmployerState         sql.NullString
	EmployerZip           sql.NullString
	Occupation            sql.NullString
	EmployerPhone         sql.NullString
	EmploymentStatus      db.NullEmploymentStatus
	EmploymentDuration    sql.NullString
	GrossMonthlyIncome    sql.NullString
	OtherMonthlyIncome    sql.NullString
	OtherIncomeSource     sql.NullString
}

func (s *ClientService) Update(ctx context.Context, in UpdateClientInput) (db.Client, error) {
	if in.Ssn.Valid {
		enc, err := encrypt(s.encryptionKey, in.Ssn.String)
		if err != nil {
			return db.Client{}, err
		}
		in.Ssn.String = enc
	}

	client, err := s.q.UpdateClient(ctx, db.UpdateClientParams{
		ID:                    in.ID,
		FirstName:             in.FirstName,
		LastName:              in.LastName,
		Phone:                 in.Phone,
		Email:                 in.Email,
		Source:                in.Source,
		Status:                in.Status,
		Notes:                 in.Notes,
		DateOfBirth:           in.DateOfBirth,
		Address:               in.Address,
		City:                  in.City,
		State:                 in.State,
		Zip:                   in.Zip,
		Ssn:                   in.Ssn,
		HousingStatus:         in.HousingStatus,
		HousingDuration:       in.HousingDuration,
		HousingMonthlyPayment: in.HousingMonthlyPayment,
		EmployerName:          in.EmployerName,
		EmployerAddress:       in.EmployerAddress,
		EmployerCity:          in.EmployerCity,
		EmployerState:         in.EmployerState,
		EmployerZip:           in.EmployerZip,
		Occupation:            in.Occupation,
		EmployerPhone:         in.EmployerPhone,
		EmploymentStatus:      in.EmploymentStatus,
		EmploymentDuration:    in.EmploymentDuration,
		GrossMonthlyIncome:    in.GrossMonthlyIncome,
		OtherMonthlyIncome:    in.OtherMonthlyIncome,
		OtherIncomeSource:     in.OtherIncomeSource,
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
