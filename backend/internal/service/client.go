package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type ClientService struct {
	q *db.Queries
}

func NewClient(q *db.Queries) *ClientService {
	return &ClientService{q: q}
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
	return s.q.CreateClient(ctx, db.CreateClientParams{
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
}

func (s *ClientService) Get(ctx context.Context, id uuid.UUID) (db.Client, error) {
	return s.q.GetClient(ctx, id)
}

func (s *ClientService) List(ctx context.Context) ([]db.Client, error) {
	return s.q.ListClients(ctx)
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
	return s.q.UpdateClient(ctx, db.UpdateClientParams{
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
}
