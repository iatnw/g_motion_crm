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

type tallyDocumentImport struct {
	SourceURL    string
	Filename     string
	DocumentType db.DocumentType
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

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after commit; ensures cleanup on all return paths

	txClient := s.client.WithTx(tx)
	txDeal := s.deal.WithTx(tx)
	txQueries := s.q.WithTx(tx)

	tallyEventID, claimed, err := claimTallyEvent(ctx, tx, payload.EventID, payload.Data.SubmissionID)
	if err != nil {
		return fmt.Errorf("claim tally event: %w", err)
	}
	if !claimed {
		_ = tx.Rollback(ctx)
		return nil
	}

	existing, lookupErr := findExistingClient(ctx, txQueries, phone, email)
	if lookupErr == nil {
		clientID = uuid.UUID(existing.ID.Bytes)
		_, err = txClient.MergeFromTally(ctx, UpdateClientInput{
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
	for _, item := range tallyDocumentImports(payload.Data.Fields) {
		if _, err = tx.Exec(ctx, `
			INSERT INTO tally_document_imports (
				tally_event_id, client_id, deal_id, source_url, file_name, document_type
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`, tallyEventID, clientID, uuid.UUID(createdDeal.ID.Bytes), item.SourceURL, item.Filename, item.DocumentType); err != nil {
			return fmt.Errorf("queue tally document import: %w", err)
		}
	}

	if err = markTallyEventProcessed(ctx, tx, payload.EventID, payload.Data.SubmissionID); err != nil {
		return fmt.Errorf("mark tally event processed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func claimTallyEvent(ctx context.Context, tx pgx.Tx, eventID, submissionID string) (int64, bool, error) {
	eventID = strings.TrimSpace(eventID)
	submissionID = strings.TrimSpace(submissionID)
	if eventID == "" && submissionID == "" {
		return 0, false, fmt.Errorf("missing event_id and submission_id")
	}

	// Try to insert a new record for this event.
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO tally_webhook_events (event_id, submission_id)
		VALUES (NULLIF($1, ''), NULLIF($2, ''))
		ON CONFLICT DO NOTHING
		RETURNING id
	`, eventID, submissionID).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if err != pgx.ErrNoRows {
		return 0, false, err
	}

	// Row already exists — check whether the previous attempt completed.
	// processed_at IS NULL means it was claimed but the processing failed; retry.
	var processedAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT id, processed_at FROM tally_webhook_events
		WHERE (event_id IS NOT DISTINCT FROM NULLIF($1, '') OR submission_id IS NOT DISTINCT FROM NULLIF($2, ''))
		LIMIT 1
	`, eventID, submissionID).Scan(&id, &processedAt)
	if err != nil {
		return 0, false, fmt.Errorf("check processed status: %w", err)
	}
	return id, !processedAt.Valid, nil
}

func markTallyEventProcessed(ctx context.Context, tx pgx.Tx, eventID, submissionID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE tally_webhook_events SET processed_at = NOW()
		WHERE (event_id IS NOT DISTINCT FROM NULLIF($1, '') OR submission_id IS NOT DISTINCT FROM NULLIF($2, ''))
	`, eventID, submissionID)
	return err
}

func (s *TallyService) StartDocumentImportScheduler(ctx context.Context) {
	go func() {
		s.retryPendingDocumentImports(ctx)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.retryPendingDocumentImports(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *TallyService) retryPendingDocumentImports(ctx context.Context) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT tally_event_id
		FROM tally_document_imports
		WHERE status = 'pending'
		   OR (status = 'processing' AND updated_at < NOW() - INTERVAL '15 minutes')
	`)
	if err != nil {
		slog.Error("tally: list pending document imports", "error", err)
		return
	}
	defer rows.Close()

	var eventIDs []int64
	for rows.Next() {
		var eventID int64
		if err := rows.Scan(&eventID); err != nil {
			slog.Error("tally: scan pending document import", "error", err)
			return
		}
		eventIDs = append(eventIDs, eventID)
	}
	if err := rows.Err(); err != nil {
		slog.Error("tally: iterate pending document imports", "error", err)
		return
	}

	for _, eventID := range eventIDs {
		if err := s.processPendingDocumentImports(ctx, eventID); err != nil {
			slog.Error("tally: retry document imports", "tally_event_id", eventID, "error", err)
		}
	}
}

func (s *TallyService) processPendingDocumentImports(ctx context.Context, tallyEventID int64) error {
	for {
		var (
			id        uuid.UUID
			clientID  uuid.UUID
			dealID    uuid.NullUUID
			sourceURL string
			filename  string
			docType   db.DocumentType
		)
		err := s.pool.QueryRow(ctx, `
			UPDATE tally_document_imports
			SET status = 'processing',
			    attempts = attempts + 1,
			    updated_at = NOW()
			WHERE id = (
				SELECT id
				FROM tally_document_imports
				WHERE tally_event_id = $1
				  AND (
				      status = 'pending'
				      OR (status = 'processing' AND updated_at < NOW() - INTERVAL '15 minutes')
				  )
				ORDER BY created_at, id
				FOR UPDATE SKIP LOCKED
				LIMIT 1
			)
			RETURNING id, client_id, deal_id, source_url, file_name, document_type
		`, tallyEventID).Scan(&id, &clientID, &dealID, &sourceURL, &filename, &docType)
		if err == pgx.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("claim tally document import: %w", err)
		}

		doc, err := s.document.UploadFromURL(ctx, sourceURL, filename, uuid.NullUUID{UUID: clientID, Valid: true}, dealID, docType)
		if err != nil {
			if _, updateErr := s.pool.Exec(ctx, `
				UPDATE tally_document_imports
				SET status = 'pending', last_error = $2, updated_at = NOW()
				WHERE id = $1
			`, id, err.Error()); updateErr != nil {
				return fmt.Errorf("upload tally document: %v; record failure: %w", err, updateErr)
			}
			return fmt.Errorf("upload tally document: %w", err)
		}

		if _, err := s.pool.Exec(ctx, `
			UPDATE tally_document_imports
			SET status = 'completed',
			    document_id = $2,
			    last_error = NULL,
			    completed_at = NOW(),
			    updated_at = NOW()
			WHERE id = $1
		`, id, uuid.UUID(doc.ID.Bytes)); err != nil {
			return fmt.Errorf("complete tally document import: %w", err)
		}
	}
}

func tallyDocumentImports(fields []TallyField) []tallyDocumentImport {
	var imports []tallyDocumentImport
	for _, field := range fields {
		if field.Type != "FILE_UPLOAD" && field.Type != "SIGNATURE" {
			continue
		}
		docType := labelToDocType(field.Label)
		for _, file := range parseFiles(field.Value) {
			if file.URL == "" {
				continue
			}
			imports = append(imports, tallyDocumentImport{
				SourceURL:    file.URL,
				Filename:     file.Name,
				DocumentType: docType,
			})
		}
	}
	return imports
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
