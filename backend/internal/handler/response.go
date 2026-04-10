package handler

import (
	"database/sql"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func ns(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

// --- Client ---

type ClientResponse struct {
	ID                    string  `json:"id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Phone                 string  `json:"phone"`
	Email                 string  `json:"email"`
	Source                *string `json:"source"`
	Status                string  `json:"status"`
	Notes                 *string `json:"notes"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
	DateOfBirth           *string `json:"date_of_birth"`
	Address               *string `json:"address"`
	City                  *string `json:"city"`
	State                 *string `json:"state"`
	Zip                   *string `json:"zip"`
	Ssn                   *string `json:"ssn"`
	HousingStatus         *string `json:"housing_status"`
	HousingDuration       *string `json:"housing_duration"`
	HousingMonthlyPayment *string `json:"housing_monthly_payment"`
	EmployerName          *string `json:"employer_name"`
	EmployerAddress       *string `json:"employer_address"`
	EmployerCity          *string `json:"employer_city"`
	EmployerState         *string `json:"employer_state"`
	EmployerZip           *string `json:"employer_zip"`
	Occupation            *string `json:"occupation"`
	EmployerPhone         *string `json:"employer_phone"`
	EmploymentStatus      *string `json:"employment_status"`
	EmploymentDuration    *string `json:"employment_duration"`
	GrossMonthlyIncome    *string `json:"gross_monthly_income"`
	OtherMonthlyIncome    *string `json:"other_monthly_income"`
	OtherIncomeSource     *string `json:"other_income_source"`
}

func toClientResponse(c db.Client) ClientResponse {
	r := ClientResponse{
		ID:                    c.ID.String(),
		FirstName:             c.FirstName,
		LastName:              c.LastName,
		Phone:                 c.Phone,
		Email:                 c.Email,
		Status:                string(c.Status),
		Notes:                 ns(c.Notes),
		CreatedAt:             c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Address:               ns(c.Address),
		City:                  ns(c.City),
		State:                 ns(c.State),
		Zip:                   ns(c.Zip),
		Ssn:                   ns(c.Ssn),
		HousingStatus:         ns(c.HousingStatus),
		HousingDuration:       ns(c.HousingDuration),
		HousingMonthlyPayment: ns(c.HousingMonthlyPayment),
		EmployerName:          ns(c.EmployerName),
		EmployerAddress:       ns(c.EmployerAddress),
		EmployerCity:          ns(c.EmployerCity),
		EmployerState:         ns(c.EmployerState),
		EmployerZip:           ns(c.EmployerZip),
		Occupation:            ns(c.Occupation),
		EmployerPhone:         ns(c.EmployerPhone),
		EmploymentDuration:    ns(c.EmploymentDuration),
		GrossMonthlyIncome:    ns(c.GrossMonthlyIncome),
		OtherMonthlyIncome:    ns(c.OtherMonthlyIncome),
		OtherIncomeSource:     ns(c.OtherIncomeSource),
	}
	if c.Source.Valid {
		r.Source = new(string(c.Source.ClientsSource))
	}
	if c.DateOfBirth.Valid {
		r.DateOfBirth = new(string(c.DateOfBirth.Time.Format("2006-01-02")))
	}
	if c.EmploymentStatus.Valid {
		r.EmploymentStatus = new(string(c.EmploymentStatus.EmploymentStatus))
	}
	return r
}

func toClientSummary(c db.Client) ClientResponse {
	r := toClientResponse(c)
	r.Ssn = nil
	return r
}

func toClientResponses(cs []db.Client) []ClientResponse {
	out := make([]ClientResponse, len(cs))
	for i, c := range cs {
		out[i] = toClientSummary(c)
	}
	return out
}

// --- Deal ---

type DealResponse struct {
	ID                       string  `json:"id"`
	ClientID                 string  `json:"client_id"`
	BrokerID                 *string `json:"broker_id"`
	DealerID                 *string `json:"dealer_id"`
	DealDate                 *string `json:"deal_date"`
	DealNumber               *string `json:"deal_number"`
	Vin                      *string `json:"vin"`
	Make                     *string `json:"make"`
	Model                    *string `json:"model"`
	TrimLevel                *string `json:"trim_level"`
	DownPayment              *string `json:"down_payment"`
	DownPaymentMethod        *string `json:"down_payment_method"`
	MonthlyPayment           *string `json:"monthly_payment"`
	BrokerFee                *string `json:"broker_fee"`
	BrokerPercent            *string `json:"broker_percent"`
	BrokerCommission         *string `json:"broker_commission"`
	BrokerCommissionStatus   *string `json:"broker_commission_status"`
	BrokerCommissionPaidDate *string `json:"broker_commission_paid_date"`
	ReferralAmount           *string `json:"referral_amount"`
	YourNet                  *string `json:"your_net"`
	BrokerFeeStatus          *string `json:"broker_fee_status"`
	BrokerFeePaidDate        *string `json:"broker_fee_paid_date"`
	ReferralStatus           *string `json:"referral_status"`
	ReferralPaidDate         *string `json:"referral_paid_date"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

func toDealResponse(d db.Deal) DealResponse {
	r := DealResponse{
		ID:               d.ID.String(),
		ClientID:         d.ClientID.String(),
		DealNumber:       ns(d.DealNumber),
		Vin:              ns(d.Vin),
		Make:             ns(d.Make),
		Model:            ns(d.Model),
		TrimLevel:        ns(d.TrimLevel),
		DownPayment:      ns(d.DownPayment),
		MonthlyPayment:   ns(d.MonthlyPayment),
		BrokerFee:        ns(d.BrokerFee),
		BrokerPercent:    ns(d.BrokerPercent),
		BrokerCommission: ns(d.BrokerCommission),
		ReferralAmount:   ns(d.ReferralAmount),
		YourNet:          ns(d.YourNet),
		CreatedAt:        d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if d.BrokerID.Valid {
		r.BrokerID = new(d.BrokerID.UUID.String())
	}
	if d.DealerID.Valid {
		r.DealerID = new(d.DealerID.UUID.String())
	}
	if d.DealDate.Valid {
		r.DealDate = new(d.DealDate.Time.Format("2006-01-02"))
	}
	if d.DownPaymentMethod.Valid {
		r.DownPaymentMethod = new(string(d.DownPaymentMethod.DownPaymentMethod))
	}
	if d.BrokerCommissionStatus.Valid {
		r.BrokerCommissionStatus = new(string(d.BrokerCommissionStatus.PaymentStatus))
	}
	if d.BrokerCommissionPaidDate.Valid {
		r.BrokerCommissionPaidDate = new(d.BrokerCommissionPaidDate.Time.Format("2006-01-02"))
	}
	if d.BrokerFeeStatus.Valid {
		r.BrokerFeeStatus = new(string(d.BrokerFeeStatus.PaymentStatus))
	}
	if d.BrokerFeePaidDate.Valid {
		r.BrokerFeePaidDate = new(d.BrokerFeePaidDate.Time.Format("2006-01-02"))
	}
	if d.ReferralStatus.Valid {
		r.ReferralStatus = new(string(d.ReferralStatus.ReferralStatus))
	}
	if d.ReferralPaidDate.Valid {
		r.ReferralPaidDate = new(d.ReferralPaidDate.Time.Format("2006-01-02"))
	}
	return r
}

func toDealResponses(ds []db.Deal) []DealResponse {
	out := make([]DealResponse, len(ds))
	for i, d := range ds {
		out[i] = toDealResponse(d)
	}
	return out
}

// --- Broker ---

type BrokerResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	Notes     *string `json:"notes"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toBrokerResponse(b db.Broker) BrokerResponse {
	return BrokerResponse{
		ID:        b.ID.String(),
		FirstName: b.FirstName,
		LastName:  b.LastName,
		Phone:     ns(b.Phone),
		Email:     ns(b.Email),
		Notes:     ns(b.Notes),
		CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toBrokerResponses(bs []db.Broker) []BrokerResponse {
	out := make([]BrokerResponse, len(bs))
	for i, b := range bs {
		out[i] = toBrokerResponse(b)
	}
	return out
}

// --- Dealer ---

type DealerResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContactPerson *string `json:"contact_person"`
	Phone         *string `json:"phone"`
	Address       *string `json:"address"`
	Notes         *string `json:"notes"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func toDealerResponse(d db.Dealer) DealerResponse {
	return DealerResponse{
		ID:            d.ID.String(),
		Name:          d.Name,
		ContactPerson: ns(d.ContactPerson),
		Phone:         ns(d.Phone),
		Address:       ns(d.Address),
		Notes:         ns(d.Notes),
		CreatedAt:     d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toDealerResponses(ds []db.Dealer) []DealerResponse {
	out := make([]DealerResponse, len(ds))
	for i, d := range ds {
		out[i] = toDealerResponse(d)
	}
	return out
}

// --- Document ---

type DocumentResponse struct {
	ID            string  `json:"id"`
	DealID        *string `json:"deal_id"`
	ClientID      *string `json:"client_id"`
	S3Key         string  `json:"s3_key"`
	FileName      string  `json:"file_name"`
	FileSizeBytes *int32  `json:"file_size_bytes"`
	DocumentType  string  `json:"document_type"`
	UploadedAt    string  `json:"uploaded_at"`
}

func toDocumentResponse(d db.Document) DocumentResponse {
	r := DocumentResponse{
		ID:           d.ID.String(),
		S3Key:        d.S3Key,
		FileName:     d.FileName,
		DocumentType: string(d.DocumentType),
		UploadedAt:   d.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if d.DealID.Valid {
		r.DealID = new(d.DealID.UUID.String())
	}
	if d.ClientID.Valid {
		r.ClientID = new(d.ClientID.UUID.String())
	}
	if d.FileSizeBytes.Valid {
		r.FileSizeBytes = &d.FileSizeBytes.Int32
	}
	return r
}

func toDocumentResponses(ds []db.Document) []DocumentResponse {
	out := make([]DocumentResponse, len(ds))
	for i, d := range ds {
		out[i] = toDocumentResponse(d)
	}
	return out
}
