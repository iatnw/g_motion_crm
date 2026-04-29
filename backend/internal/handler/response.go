package handler

import (
	"math"
	"math/big"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

// --- Client ---

type ClientResponse struct {
	ID                    string   `json:"id"`
	FirstName             string   `json:"first_name"`
	LastName              string   `json:"last_name"`
	Phone                 string   `json:"phone"`
	Email                 string   `json:"email"`
	Source                *string  `json:"source"`
	Status                string   `json:"status"`
	Notes                 *string  `json:"notes"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
	DateOfBirth           *string  `json:"date_of_birth"`
	Address               *string  `json:"address"`
	City                  *string  `json:"city"`
	State                 *string  `json:"state"`
	Zip                   *string  `json:"zip"`
	Ssn                   *string  `json:"ssn"`
	HousingStatus         *string  `json:"housing_status"`
	HousingDuration       *string  `json:"housing_duration"`
	HousingMonthlyPayment *float64 `json:"housing_monthly_payment"`
	EmployerName          *string  `json:"employer_name"`
	EmployerAddress       *string  `json:"employer_address"`
	EmployerCity          *string  `json:"employer_city"`
	EmployerState         *string  `json:"employer_state"`
	EmployerZip           *string  `json:"employer_zip"`
	Occupation            *string  `json:"occupation"`
	EmployerPhone         *string  `json:"employer_phone"`
	EmploymentStatus      *string  `json:"employment_status"`
	EmploymentDuration    *string  `json:"employment_duration"`
	GrossMonthlyIncome    *float64 `json:"gross_monthly_income"`
	OtherMonthlyIncome    *float64 `json:"other_monthly_income"`
	OtherIncomeSource     *string  `json:"other_income_source"`
}

func toClientResponse(c db.Client) ClientResponse {
	r := ClientResponse{
		ID:                    pgUUID(c.ID),
		FirstName:             c.FirstName,
		LastName:              c.LastName,
		Phone:                 c.Phone,
		Email:                 c.Email,
		Status:                string(c.Status),
		Notes:                 pgText(c.Notes),
		CreatedAt:             pgTimestamp(c.CreatedAt),
		UpdatedAt:             pgTimestamp(c.UpdatedAt),
		DateOfBirth:           pgDate(c.DateOfBirth),
		Address:               pgText(c.Address),
		City:                  pgText(c.City),
		State:                 pgText(c.State),
		Zip:                   pgText(c.Zip),
		Ssn:                   pgText(c.Ssn),
		HousingStatus:         pgText(c.HousingStatus),
		HousingDuration:       pgText(c.HousingDuration),
		HousingMonthlyPayment: pgNumeric(c.HousingMonthlyPayment),
		EmployerName:          pgText(c.EmployerName),
		EmployerAddress:       pgText(c.EmployerAddress),
		EmployerCity:          pgText(c.EmployerCity),
		EmployerState:         pgText(c.EmployerState),
		EmployerZip:           pgText(c.EmployerZip),
		Occupation:            pgText(c.Occupation),
		EmployerPhone:         pgText(c.EmployerPhone),
		EmploymentDuration:    pgText(c.EmploymentDuration),
		GrossMonthlyIncome:    pgNumeric(c.GrossMonthlyIncome),
		OtherMonthlyIncome:    pgNumeric(c.OtherMonthlyIncome),
		OtherIncomeSource:     pgText(c.OtherIncomeSource),
	}
	if c.Source.Valid {
		s := string(c.Source.ClientsSource)
		r.Source = &s
	}
	if c.EmploymentStatus.Valid {
		s := string(c.EmploymentStatus.EmploymentStatus)
		r.EmploymentStatus = &s
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
	ID                       string   `json:"id"`
	ClientID                 string   `json:"client_id"`
	BrokerID                 *string  `json:"broker_id"`
	DealerID                 *string  `json:"dealer_id"`
	DealDate                 *string  `json:"deal_date"`
	DealNumber               *string  `json:"deal_number"`
	Vin                      *string  `json:"vin"`
	Make                     *string  `json:"make"`
	Model                    *string  `json:"model"`
	TrimLevel                *string  `json:"trim_level"`
	DownPayment              *float64 `json:"down_payment"`
	DownPaymentMethod        *string  `json:"down_payment_method"`
	MonthlyPayment           *float64 `json:"monthly_payment"`
	BrokerFee                *float64 `json:"broker_fee"`
	BrokerPercent            *float64 `json:"broker_percent"`
	BrokerCommission         *float64 `json:"broker_commission"`
	BrokerCommissionStatus   *string  `json:"broker_commission_status"`
	BrokerCommissionPaidDate *string  `json:"broker_commission_paid_date"`
	YourNet                  *float64 `json:"your_net"`
	BrokerFeeStatus          *string  `json:"broker_fee_status"`
	BrokerFeePaidDate        *string  `json:"broker_fee_paid_date"`
	DealStatus               *string  `json:"deal_status"`
	CollectAmount            *float64 `json:"collect_amount"`
	DocFee                   *float64 `json:"doc_fee"`
	BrokerPaymentMethod      *string  `json:"broker_payment_method"`
	Notes                    *string  `json:"notes"`
	CreatedAt                string   `json:"created_at"`
	UpdatedAt                string   `json:"updated_at"`
}

func toDealResponse(d db.Deal) DealResponse {
	r := DealResponse{
		ID:                       pgUUID(d.ID),
		ClientID:                 pgUUID(d.ClientID),
		BrokerID:                 pgNullUUID(d.BrokerID),
		DealerID:                 pgNullUUID(d.DealerID),
		DealDate:                 pgDate(d.DealDate),
		DealNumber:               pgText(d.DealNumber),
		Vin:                      pgText(d.Vin),
		Make:                     pgText(d.Make),
		Model:                    pgText(d.Model),
		TrimLevel:                pgText(d.TrimLevel),
		DownPayment:              pgNumeric(d.DownPayment),
		MonthlyPayment:           pgNumeric(d.MonthlyPayment),
		BrokerFee:                pgNumeric(d.BrokerFee),
		BrokerPercent:            pgNumeric(d.BrokerPercent),
		BrokerCommission:         pgNumeric(d.BrokerCommission),
		BrokerCommissionPaidDate: pgDate(d.BrokerCommissionPaidDate),
		YourNet:                  pgNumeric(d.YourNet),
		BrokerFeePaidDate:        pgDate(d.BrokerFeePaidDate),
		CollectAmount:            pgNumeric(d.CollectAmount),
		DocFee:                   pgNumeric(d.DocFee),
		Notes:                    pgText(d.Notes),
		CreatedAt:                pgTimestamp(d.CreatedAt),
		UpdatedAt:                pgTimestamp(d.UpdatedAt),
	}
	if d.DownPaymentMethod.Valid {
		s := string(d.DownPaymentMethod.DownPaymentMethod)
		r.DownPaymentMethod = &s
	}
	if d.BrokerCommissionStatus.Valid {
		s := string(d.BrokerCommissionStatus.PaymentStatus)
		r.BrokerCommissionStatus = &s
	}
	if d.BrokerFeeStatus.Valid {
		s := string(d.BrokerFeeStatus.PaymentStatus)
		r.BrokerFeeStatus = &s
	}
	if d.DealStatus.Valid {
		s := string(d.DealStatus.DealStatus)
		r.DealStatus = &s
	}
	if d.BrokerPaymentMethod.Valid {
		s := string(d.BrokerPaymentMethod.BrokerPaymentMethod)
		r.BrokerPaymentMethod = &s
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
		ID:        pgUUID(b.ID),
		FirstName: b.FirstName,
		LastName:  b.LastName,
		Phone:     pgText(b.Phone),
		Email:     pgText(b.Email),
		Notes:     pgText(b.Notes),
		CreatedAt: pgTimestamp(b.CreatedAt),
		UpdatedAt: pgTimestamp(b.UpdatedAt),
	}
}

func toBrokerResponses(bs []db.Broker) []BrokerResponse {
	out := make([]BrokerResponse, len(bs))
	for i, b := range bs {
		out[i] = toBrokerResponse(b)
	}
	return out
}

type BrokerStatResponse struct {
	ID                    string   `json:"id"`
	TotalDeals            int64    `json:"total_deals"`
	CurrentMonthDeals     int64    `json:"current_month_deals"`
	TotalBrokerFee        *float64 `json:"total_broker_fee"`
	TotalBrokerCommission *float64 `json:"total_broker_commission"`
	BrokerCleanIncome     *float64 `json:"broker_clean_income"`
	AvgBrokerFeePerDeal   *float64 `json:"avg_broker_fee_per_deal"`
	AvgDealsPerMonth      *float64 `json:"avg_deals_per_month"`
}

func toBrokerStatResponse(b db.BrokerStat) BrokerStatResponse {
	return BrokerStatResponse{
		ID:                    pgUUID(b.ID),
		TotalDeals:            b.TotalDeals,
		CurrentMonthDeals:     b.CurrentMonthDeals,
		TotalBrokerFee:        pgNumeric(b.TotalBrokerFee),
		TotalBrokerCommission: pgNumeric(b.TotalBrokerCommission),
		BrokerCleanIncome:     pgNumeric(b.BrokerCleanIncome),
		AvgBrokerFeePerDeal:   pgNumeric(b.AvgBrokerFeePerDeal),
		AvgDealsPerMonth:      pgNumeric(b.AvgDealsPerMonth),
	}
}

// --- Dealer ---

type DealerResponse struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	ContactPerson      *string `json:"contact_person"`
	Phone              *string `json:"phone"`
	Address            *string `json:"address"`
	Notes              *string `json:"notes"`
	ContactPersonEmail *string `json:"contact_person_email"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

func toDealerResponse(d db.Dealer) DealerResponse {
	return DealerResponse{
		ID:                 pgUUID(d.ID),
		Name:               d.Name,
		ContactPerson:      pgText(d.ContactPerson),
		Phone:              pgText(d.Phone),
		Address:            pgText(d.Address),
		Notes:              pgText(d.Notes),
		ContactPersonEmail: pgText(d.ContactPersonEmail),
		CreatedAt:          pgTimestamp(d.CreatedAt),
		UpdatedAt:          pgTimestamp(d.UpdatedAt),
	}
}

func toDealerResponses(ds []db.Dealer) []DealerResponse {
	out := make([]DealerResponse, len(ds))
	for i, d := range ds {
		out[i] = toDealerResponse(d)
	}
	return out
}

type DealerStatResponse struct {
	ID               string   `json:"id"`
	DealCount        int64    `json:"deal_count"`
	TotalBrokerFee   *float64 `json:"total_broker_fee"`
	AvgDaysToPayment *float64 `json:"avg_days_to_payment"`
}

func toDealerStatResponse(d db.DealerStat) DealerStatResponse {
	return DealerStatResponse{
		ID:               pgUUID(d.ID),
		DealCount:        d.DealCount,
		TotalBrokerFee:   pgNumeric(d.TotalBrokerFee),
		AvgDaysToPayment: pgNumeric(d.AvgDaysToPayment),
	}
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
	return DocumentResponse{
		ID:            pgUUID(d.ID),
		DealID:        pgNullUUID(d.DealID),
		ClientID:      pgNullUUID(d.ClientID),
		S3Key:         d.S3Key,
		FileName:      d.FileName,
		FileSizeBytes: pgInt4(d.FileSizeBytes),
		DocumentType:  string(d.DocumentType),
		UploadedAt:    pgTimestamp(d.UploadedAt),
	}
}

func toDocumentResponses(ds []db.Document) []DocumentResponse {
	out := make([]DocumentResponse, len(ds))
	for i, d := range ds {
		out[i] = toDocumentResponse(d)
	}
	return out
}

// --- Referral ---

type ReferralResponse struct {
	ID        string   `json:"id"`
	DealID    string   `json:"deal_id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     *string  `json:"phone"`
	Amount    *float64 `json:"amount"`
	Status    string   `json:"status"`
	PaidDate  *string  `json:"paid_date"`
	Notes     *string  `json:"notes"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func toReferralResponse(r db.Referral) ReferralResponse {
	return ReferralResponse{
		ID:        pgUUID(r.ID),
		DealID:    pgUUID(r.DealID),
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Phone:     pgText(r.Phone),
		Amount:    pgNumeric(r.Amount),
		Status:    string(r.Status),
		PaidDate:  pgDate(r.PaidDate),
		Notes:     pgText(r.Notes),
		CreatedAt: pgTimestamp(r.CreatedAt),
		UpdatedAt: pgTimestamp(r.UpdatedAt),
	}
}

func toReferralResponses(rs []db.Referral) []ReferralResponse {
	out := make([]ReferralResponse, len(rs))
	for i, r := range rs {
		out[i] = toReferralResponse(r)
	}
	return out
}

// --- pgtype helpers ---

func pgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgUUID(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}

func pgNullUUID(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func pgTimestamp(t pgtype.Timestamptz) string {
	return t.Time.Format("2006-01-02T15:04:05Z07:00")
}

func pgDate(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func pgNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid || n.NaN || n.Int == nil {
		return nil
	}
	f, _ := new(big.Float).SetInt(n.Int).Float64()
	if n.Exp != 0 {
		f *= math.Pow10(int(n.Exp))
	}
	return &f
}

func pgInt4(n pgtype.Int4) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}
