package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

// TallyService handles incoming Tally webhook submissions.
type TallyService struct {
	client   *ClientService
	document *DocumentService
	deal     *DealService
	q        *db.Queries
	pool     *pgxpool.Pool
}

func NewTally(client *ClientService, document *DocumentService, deal *DealService, q *db.Queries, pool *pgxpool.Pool) *TallyService {
	return &TallyService{client: client, document: document, deal: deal, q: q, pool: pool}
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
	seen   map[string]int
}

func newParser(fields []TallyField) *fieldParser {
	return &fieldParser{fields: fields, seen: make(map[string]int)}
}

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

func (p *fieldParser) nullStr(label string) *string {
	v := p.str(label)
	if v == "" {
		return nil
	}
	return &v
}

func (p *fieldParser) nullDate(label string) *time.Time {
	v := p.str(label)
	if v == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "01/02/2006", "1/2/2006"} {
		t, err := time.Parse(layout, v)
		if err == nil {
			return &t
		}
	}
	return nil
}

func (p *fieldParser) nullFloat(label string) *float64 {
	v := p.str(label)
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
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

	firstName := strings.TrimSpace(p.str("First legal name (as on Driver License)"))
	lastName := strings.TrimSpace(p.str("Last legal name"))
	if firstName == "" {
		// fallback for older form version with a single combined field
		fullName := p.str("Full legal name")
		var err error
		firstName, lastName, err = splitName(fullName)
		if err != nil {
			return fmt.Errorf("missing required field: first name")
		}
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
	housingMonthlyPayment := p.nullFloat("Monthly Payment")

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
	grossMonthlyIncome := p.nullFloat("Gross Monthly Income")
	otherMonthlyIncome := p.nullFloat("Other Monthly Income")
	otherIncomeSource := p.nullStr("Other Income Source")

	vehicleOfInterest := p.str("Vehicle of Interest")
	vehicleMake, vehicleModel := splitVehicle(vehicleOfInterest)

	var clientID uuid.UUID
	var createdDeal db.Deal
	var dealCreated bool

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after commit; ensures cleanup on all return paths

	txClient := s.client.WithTx(tx)
	txDeal := s.deal.WithTx(tx)
	txQueries := s.q.WithTx(tx)

	claimed, err := claimTallyEvent(ctx, tx, payload.EventID, payload.Data.SubmissionID)
	if err != nil {
		return fmt.Errorf("claim tally event: %w", err)
	}
	if !claimed {
		return nil
	}

	existing, lookupErr := findExistingClient(ctx, txQueries, phone, email)
	if lookupErr == nil {
		clientID = uuid.UUID(existing.ID.Bytes)
		_, err = txClient.Update(ctx, UpdateClientInput{
			ID:                    clientID,
			FirstName:             firstName,
			LastName:              lastName,
			Phone:                 strPtr(phone),
			Email:                 strPtr(email),
			Source:                existing.Source,
			Status:                existing.Status,
			Notes:                 fromText(existing.Notes),
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
			Phone:                 strPtr(phone),
			Email:                 strPtr(email),
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
		clientID = uuid.UUID(created.ID.Bytes)
	}

	createdDeal, err = txDeal.Create(ctx, CreateDealInput{
		ClientID: clientID,
		Make:     strPtr(vehicleMake),
		Model:    strPtr(vehicleModel),
	})
	if err != nil {
		return fmt.Errorf("create deal: %w", err)
	}
	dealCreated = true

	if err = markTallyEventProcessed(ctx, tx, payload.EventID, payload.Data.SubmissionID); err != nil {
		return fmt.Errorf("mark tally event processed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Document uploads happen after commit — S3 is not transactional.
	clientNullUUID := uuid.NullUUID{UUID: clientID, Valid: true}
	dealNullUUID := uuid.NullUUID{}
	if dealCreated {
		dealNullUUID = uuid.NullUUID{UUID: uuid.UUID(createdDeal.ID.Bytes), Valid: true}
	}

	for _, field := range payload.Data.Fields {
		if field.Type != "FILE_UPLOAD" && field.Type != "SIGNATURE" {
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

func claimTallyEvent(ctx context.Context, tx pgx.Tx, eventID, submissionID string) (bool, error) {
	eventID = strings.TrimSpace(eventID)
	submissionID = strings.TrimSpace(submissionID)
	if eventID == "" && submissionID == "" {
		return false, fmt.Errorf("missing event_id and submission_id")
	}

	// Try to insert a new record for this event.
	var claimed bool
	err := tx.QueryRow(ctx, `
		INSERT INTO tally_webhook_events (event_id, submission_id)
		VALUES (NULLIF($1, ''), NULLIF($2, ''))
		ON CONFLICT DO NOTHING
		RETURNING true
	`, eventID, submissionID).Scan(&claimed)
	if err == nil {
		return true, nil
	}
	if err != pgx.ErrNoRows {
		return false, err
	}

	// Row already exists — check whether the previous attempt completed.
	// processed_at IS NULL means it was claimed but the processing failed; retry.
	var processedAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT processed_at FROM tally_webhook_events
		WHERE (event_id IS NOT DISTINCT FROM NULLIF($1, '') OR submission_id IS NOT DISTINCT FROM NULLIF($2, ''))
		LIMIT 1
	`, eventID, submissionID).Scan(&processedAt)
	if err != nil {
		return false, fmt.Errorf("check processed status: %w", err)
	}
	return !processedAt.Valid, nil
}

func markTallyEventProcessed(ctx context.Context, tx pgx.Tx, eventID, submissionID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE tally_webhook_events SET processed_at = NOW()
		WHERE (event_id IS NOT DISTINCT FROM NULLIF($1, '') OR submission_id IS NOT DISTINCT FROM NULLIF($2, ''))
	`, eventID, submissionID)
	return err
}

func findExistingClient(ctx context.Context, q *db.Queries, phone, email string) (db.Client, error) {
	if phone != "" {
		c, err := q.GetClientByPhone(ctx, toText(strPtr(phone)))
		if err == nil {
			return c, nil
		}
	}
	if email != "" {
		c, err := q.GetClientByEmail(ctx, email)
		if err == nil {
			return c, nil
		}
	}
	return db.Client{}, fmt.Errorf("not found")
}

func splitName(full string) (string, string, error) {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("full legal name is required")
	}
	if len(parts) == 1 {
		return parts[0], parts[0], nil
	}
	return parts[0], strings.Join(parts[1:], " "), nil
}

func splitVehicle(s string) (vehicleMake, vehicleModel string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	parts := strings.Fields(s)
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
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func labelToDocType(label string) db.DocumentType {
	switch strings.ToLower(label) {
	case "driver's license", "drivers license", "driver license":
		return db.DocumentTypeDriverLicense
	case "insurance", "insurance card":
		return db.DocumentTypeInsurance
	case "broker agreement":
		return db.DocumentTypeBrokerAgreement
	case "dealer docs":
		return db.DocumentTypeDealerDocs
	case "check copy":
		return db.DocumentTypeCheckCopy
	case "credit app", "credit application":
		return db.DocumentTypeCreditAppPdf
	case "signature":
		return db.DocumentTypeOther
	default:
		return db.DocumentTypeOther
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
