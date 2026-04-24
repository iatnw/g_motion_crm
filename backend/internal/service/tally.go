package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

// TallyService handles incoming Tally webhook submissions.
type TallyService struct {
	client   *ClientService
	document *DocumentService
	deal     *DealService
	q        *db.Queries
	db       *sql.DB
}

func NewTally(client *ClientService, document *DocumentService, deal *DealService, q *db.Queries, db *sql.DB) *TallyService {
	return &TallyService{client: client, document: document, deal: deal, q: q, db: db}
}

// --- Tally webhook payload types ---

type TallyWebhook struct {
	EventID   string    `json:"eventId"`
	EventType string    `json:"eventType"`
	CreatedAt time.Time `json:"createdAt"`
	Data      TallyData `json:"data"`
}

type TallyData struct {
	ResponseID   string       `json:"responseId"`
	SubmissionID string       `json:"submissionId"`
	FormID       string       `json:"formId"`
	FormName     string       `json:"formName"`
	CreatedAt    time.Time    `json:"createdAt"`
	Fields       []TallyField `json:"fields"`
}

type TallyField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type TallyFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

// fieldParser provides positional-aware field access for forms with duplicate labels.
type fieldParser struct {
	fields []TallyField
	// counts how many times we've seen each normalized label
	seen map[string]int
}

func newParser(fields []TallyField) *fieldParser {
	return &fieldParser{fields: fields, seen: make(map[string]int)}
}

// str returns the value of the nth occurrence (0-based) of a label.
func (p *fieldParser) str(label string) string {
	key := strings.ToLower(strings.TrimSpace(label))
	count := 0
	for _, f := range p.fields {
		if strings.ToLower(strings.TrimSpace(f.Label)) == key {
			if count == p.seen[key] {
				p.seen[key]++
				return fieldString(f.Value)
			}
			count++
		}
	}
	return ""
}

func (p *fieldParser) nullStr(label string) sql.NullString {
	v := p.str(label)
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}

func (p *fieldParser) nullDate(label string) sql.NullTime {
	v := p.str(label)
	if v == "" {
		return sql.NullTime{}
	}
	for _, layout := range []string{"2006-01-02", "01/02/2006", "1/2/2006"} {
		t, err := time.Parse(layout, v)
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	return sql.NullTime{}
}

func (p *fieldParser) nullEmployment(label string) db.NullEmploymentStatus {
	v := strings.ToLower(p.str(label))
	switch {
	case strings.Contains(v, "self"):
		return db.NullEmploymentStatus{EmploymentStatus: db.EmploymentStatusSelfEmployed, Valid: true}
	case strings.Contains(v, "full"):
		return db.NullEmploymentStatus{EmploymentStatus: db.EmploymentStatusFullTime, Valid: true}
	case strings.Contains(v, "part"):
		return db.NullEmploymentStatus{EmploymentStatus: db.EmploymentStatusPartTime, Valid: true}
	}
	return db.NullEmploymentStatus{}
}

func fieldString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

// ProcessSubmission creates or updates a client from a Tally form submission.
func (s *TallyService) ProcessSubmission(ctx context.Context, payload TallyWebhook) error {
	p := newParser(payload.Data.Fields)

	fullName := p.str("Full legal name")
	firstName, lastName, err := splitName(fullName)
	if err != nil {
		return fmt.Errorf("missing required field: full legal name")
	}
	if firstName == "" {
		return fmt.Errorf("missing required field: full legal name")
	}

	phone := p.str("Phone Number")
	email := p.str("Email")

	if phone == "" && email == "" {
		return fmt.Errorf("submission missing both phone and email")
	}

	// Contact info section (first city/state/zip occurrence)
	dob := p.nullDate("Date of Birth")
	address := p.nullStr("Address")
	city := p.nullStr("City")
	state := p.nullStr("State")
	zip := p.nullStr("Zip")
	ssn := p.nullStr("Social Security Number")
	housingStatus := p.nullStr("Housing Status")
	housingDuration := p.nullStr("How Long?")
	housingMonthlyPayment := p.nullStr("Monthly Payment")

	// Employment section (second city/state/zip occurrence)
	employerName := p.nullStr("Employer Name")
	employerAddress := p.nullStr("Employer Address")
	employerCity := p.nullStr("City")
	employerState := p.nullStr("State")
	employerZip := p.nullStr("Zip")
	occupation := p.nullStr("Occupation")
	employerPhone := p.nullStr("Employer Phone Number")
	employmentStatus := p.nullEmployment("Employment Status")
	employmentDuration := p.nullStr("How Long?")
	grossMonthlyIncome := p.nullStr("Gross Monthly Income")
	otherMonthlyIncome := p.nullStr("Other Monthly Income")
	otherIncomeSource := p.nullStr("Other Income Source")

	vehicleOfInterest := p.str("Vehicle of Interest")
	vehicleMake, vehicleModel := splitVehicle(vehicleOfInterest)

	// All DB writes (client + deal) run inside a single transaction so that
	// a failure at any step leaves no orphaned records.
	var clientID uuid.UUID
	var createdDeal db.Deal
	var dealCreated bool

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	txClient := s.client.WithTx(tx)
	txDeal := s.deal.WithTx(tx)

	existing, lookupErr := s.findExistingClient(ctx, phone, email)
	if lookupErr == nil {
		clientID = existing.ID
		_, err = txClient.Update(ctx, UpdateClientInput{
			ID:                    clientID,
			FirstName:             firstName,
			LastName:              lastName,
			Phone:                 phone,
			Email:                 email,
			Source:                existing.Source,
			Status:                existing.Status,
			Notes:                 existing.Notes,
			DateOfBirth:           dob,
			Address:               address,
			City:                  city,
			State:                 state,
			Zip:                   zip,
			Ssn:                   ssn,
			HousingStatus:         housingStatus,
			HousingDuration:       housingDuration,
			HousingMonthlyPayment: housingMonthlyPayment,
			EmployerName:          employerName,
			EmployerAddress:       employerAddress,
			EmployerCity:          employerCity,
			EmployerState:         employerState,
			EmployerZip:           employerZip,
			Occupation:            occupation,
			EmployerPhone:         employerPhone,
			EmploymentStatus:      employmentStatus,
			EmploymentDuration:    employmentDuration,
			GrossMonthlyIncome:    grossMonthlyIncome,
			OtherMonthlyIncome:    otherMonthlyIncome,
			OtherIncomeSource:     otherIncomeSource,
		})
		if err != nil {
			return fmt.Errorf("update client: %w", err)
		}
	} else {
		var created db.Client
		created, err = txClient.Create(ctx, CreateClientInput{
			FirstName:             firstName,
			LastName:              lastName,
			Phone:                 phone,
			Email:                 email,
			Source:                db.NullClientsSource{ClientsSource: db.ClientsSourceOrganic, Valid: true},
			Status:                db.ClientStatusActive,
			DateOfBirth:           dob,
			Address:               address,
			City:                  city,
			State:                 state,
			Zip:                   zip,
			Ssn:                   ssn,
			HousingStatus:         housingStatus,
			HousingDuration:       housingDuration,
			HousingMonthlyPayment: housingMonthlyPayment,
			EmployerName:          employerName,
			EmployerAddress:       employerAddress,
			EmployerCity:          employerCity,
			EmployerState:         employerState,
			EmployerZip:           employerZip,
			Occupation:            occupation,
			EmployerPhone:         employerPhone,
			EmploymentStatus:      employmentStatus,
			EmploymentDuration:    employmentDuration,
			GrossMonthlyIncome:    grossMonthlyIncome,
			OtherMonthlyIncome:    otherMonthlyIncome,
			OtherIncomeSource:     otherIncomeSource,
		})
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}
		clientID = created.ID
	}

	createdDeal, err = txDeal.Create(ctx, CreateDealInput{
		ClientID: clientID,
		Make:     sql.NullString{String: vehicleMake, Valid: vehicleMake != ""},
		Model:    sql.NullString{String: vehicleModel, Valid: vehicleModel != ""},
	})
	if err != nil {
		return fmt.Errorf("create deal: %w", err)
	}
	dealCreated = true

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Document uploads happen after commit — S3 is not transactional.
	// Failures are logged but do not affect the already-committed client/deal.
	clientNullUUID := uuid.NullUUID{UUID: clientID, Valid: true}
	dealNullUUID := uuid.NullUUID{}
	if dealCreated {
		dealNullUUID = uuid.NullUUID{UUID: createdDeal.ID, Valid: true}
	}

	for _, field := range payload.Data.Fields {
		if field.Type != "FILE_UPLOAD" {
			continue
		}
		files := parseFiles(field.Value)
		docType := labelToDocType(field.Label)
		for _, f := range files {
			if f.URL == "" {
				continue
			}
			_, err := s.document.UploadFromURL(ctx, f.URL, f.Name, clientNullUUID, dealNullUUID, docType)
			if err != nil {
				slog.Error("tally: failed to upload document", "file", f.Name, "client_id", clientID, "error", err)
			}
		}
	}

	return nil
}

func (s *TallyService) findExistingClient(ctx context.Context, phone, email string) (db.Client, error) {
	if phone != "" {
		c, err := s.q.GetClientByPhone(ctx, phone)
		if err == nil {
			return c, nil
		}
	}
	if email != "" {
		c, err := s.q.GetClientByEmail(ctx, email)
		if err == nil {
			return c, nil
		}
	}
	return db.Client{}, fmt.Errorf("not found")
}

func splitName(full string) (string, string, error) {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("firstname and lastname required fields")
	}
	if len(parts) == 1 {
		return parts[0], "", fmt.Errorf("lastname is a required field")
	}
	return parts[0], strings.Join(parts[1:], " "), nil
}

// splitVehicle splits "Toyota Camry" → ("Toyota", "Camry"),
// "2024 Toyota Camry" → ("Toyota", "2024 Camry"), or "Toyota" → ("Toyota", "").
// Year prefixes (4-digit numbers) are moved to the model portion.
func splitVehicle(s string) (vehicleMake, vehicleModel string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	parts := strings.Fields(s)
	// If first token looks like a year (4 digits), skip it into model
	start := 0
	yearPrefix := ""
	if len(parts[0]) == 4 {
		allDigits := true
		for _, c := range parts[0] {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			yearPrefix = parts[0] + " "
			start = 1
		}
	}
	if start >= len(parts) {
		return "", ""
	}
	vehicleMake = parts[start]
	if start+1 < len(parts) {
		vehicleModel = yearPrefix + strings.Join(parts[start+1:], " ")
	}
	return vehicleMake, vehicleModel
}

func parseFiles(value any) []TallyFile {
	if value == nil {
		return nil
	}
	arr, ok := value.([]any)
	if !ok {
		return nil
	}
	var files []TallyFile
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		f := TallyFile{
			Name: stringFromMap(m, "name"),
			URL:  stringFromMap(m, "url"),
		}
		if f.URL != "" {
			files = append(files, f)
		}
	}
	return files
}

func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func labelToDocType(label string) db.DocumentType {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "driver") || strings.Contains(lower, "license"):
		return db.DocumentTypeDriverLicense
	case strings.Contains(lower, "insurance"):
		return db.DocumentTypeInsurance
	default:
		return db.DocumentTypeOther
	}
}
