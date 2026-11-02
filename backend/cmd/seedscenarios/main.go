// seedscenarios adds repeatable, fictional CRM workflow fixtures without resetting data.
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
	"github.com/umbrllacrp/g_motion_crm/internal/storage"
)

func ptr[T any](v T) *T          { return &v }
func id(v pgtype.UUID) uuid.UUID { return uuid.UUID(v.Bytes) }
func text(v string) pgtype.Text  { return pgtype.Text{String: v, Valid: true} }

const marker = "local-demo-v1"

type scenario struct {
	key, first, last, make, model                             string
	age, months                                               int
	cash, paid, cancelled, vinless, direct, commissionPending bool
}

var scenarios = []scenario{
	{key: "cash-paid", first: "Avery", last: "Brooks", make: "Toyota", model: "RAV4", age: 5, cash: true, paid: true},
	{key: "cash-pending", first: "Jordan", last: "Hayes", make: "Honda", model: "CR-V", age: 15, cash: true},
	{key: "cash-cancelled", first: "Taylor", last: "Reed", make: "Ford", model: "Explorer", age: 12, cash: true, cancelled: true},
	{key: "lease-active", first: "Morgan", last: "Ellis", make: "BMW", model: "X3", age: 90, months: 36, paid: true},
	{key: "lease-ending-soon", first: "Riley", last: "Stone", make: "Lexus", model: "NX", age: 0, months: 36, paid: true},
	{key: "lease-ended", first: "Casey", last: "Lane", make: "Audi", model: "A4", age: 0, months: 24, paid: true},
	{key: "lease-long-term", first: "Jamie", last: "Cole", make: "Mercedes-Benz", model: "GLC", age: 180, months: 48, paid: true},
	{key: "new-lead", first: "Quinn", last: "Harper", make: "Hyundai", model: "Tucson", age: 1, vinless: true},
	{key: "cosigner-intake", first: "Drew", last: "Bennett", make: "Kia", model: "Sportage", age: 7, vinless: true},
	{key: "cosigned-finance", first: "Alexis", last: "Parker", make: "Kia", model: "Sportage", age: 45},
	{key: "overdue-direct", first: "Cameron", last: "Foster", make: "Chevrolet", model: "Tahoe", age: 65, direct: true},
	{key: "finance-cancelled", first: "Dakota", last: "Blake", make: "Nissan", model: "Rogue", age: 10, cancelled: true},
	{key: "commission-unpaid", first: "Skyler", last: "Morgan", make: "Subaru", model: "Outback", age: 8, paid: true, commissionPending: true},
	{key: "zero-fee-direct", first: "Reese", last: "Sutton", make: "Mazda", model: "CX-5", age: 2, paid: true, direct: true},
	{key: "new-incomplete", first: "Rowan", last: "West", age: 0, vinless: true, direct: true},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.App.Env != "development" || cfg.AWS.S3Endpoint == "" {
		log.Fatal("requires development environment and local S3 endpoint")
	}
	key, err := hex.DecodeString(cfg.App.EncryptionKey)
	if err != nil || len(key) != 32 {
		log.Fatal("invalid encryption key")
	}
	pool, err := pgxpool.New(ctx, cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := seedBusiness(ctx, pool, key); err != nil {
		log.Fatal(err)
	}
	q := db.New(pool)
	s3, err := storage.NewS3Client(cfg.AWS)
	if err != nil {
		log.Fatal(err)
	}
	documents := service.NewDocument(q, s3, key)
	if err := seedDocuments(ctx, pool, documents); err != nil {
		log.Fatal(err)
	}
	tally := service.NewTally(service.NewClient(q, key), documents, service.NewDeal(q, pool), q, pool)
	submission := service.TallyWebhook{EventID: marker + "-event", EventType: "FORM_RESPONSE", CreatedAt: time.Now(), Data: service.TallyData{
		SubmissionID: marker + "-submission", FormID: "local-demo", FormName: "Local demo intake", CreatedAt: time.Now(), Fields: []service.TallyField{
			{Label: "Full legal name", Value: "Finley DemoIntake"}, {Label: "Email", Value: "tally-intake@example.test"},
			{Label: "Phone Number", Value: "202-555-0199"}, {Label: "Vehicle of Interest", Value: "Toyota Corolla"},
		},
	}}
	if err := tally.ProcessSubmission(ctx, submission); err != nil {
		log.Fatal(err)
	}
	// Replaying the submission should be a no-op, exercising webhook deduplication.
	if err := tally.ProcessSubmission(ctx, submission); err != nil {
		log.Fatal(err)
	}
	log.Println("Scenario data, sample documents, and deduplicated Tally intake are ready")
}

func seedBusiness(ctx context.Context, pool *pgxpool.Pool, key []byte) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, marker); err != nil {
		return err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM clients WHERE email=$1)`, marker+"-00@example.test").Scan(&exists); err != nil {
		return err
	}
	if exists {
		log.Println("Business scenarios already exist; preserving them")
		return nil
	}
	q := db.New(tx)
	brokers := service.NewBroker(q)
	dealers := service.NewDealer(q)
	clients := service.NewClient(q, key)
	deals := service.NewDeal(q, pool)
	refs := service.NewReferral(q)
	broker, err := brokers.Create(ctx, service.CreateBrokerInput{FirstName: "Demo", LastName: "Partner", Email: ptr("partner@example.test"), Phone: ptr("202-555-0118"), Notes: ptr("Fictional local workflow samples")})
	if err != nil {
		return err
	}
	if _, err = brokers.Create(ctx, service.CreateBrokerInput{FirstName: "Demo", LastName: "NoDeals", Email: ptr("no-deals@example.test"), Notes: ptr("Empty broker stats example")}); err != nil {
		return err
	}
	dealer, err := dealers.Create(ctx, service.CreateDealerInput{Name: "Demo Auto Center", ContactPerson: ptr("Sam Demo"), ContactPersonEmail: ptr("dealer@example.test"), Phone: ptr("202-555-0119"), Address: ptr("100 Example Avenue, Los Angeles, CA 90001"), Notes: ptr("Fictional local workflow samples")})
	if err != nil {
		return err
	}
	if _, err = dealers.Create(ctx, service.CreateDealerInput{Name: "Demo Dealer - No Deals", Notes: ptr("Empty dealer stats example")}); err != nil {
		return err
	}
	var createdClients []db.Client
	var createdDeals []db.Deal
	sources := []db.ClientsSource{db.ClientsSourceInstagram, db.ClientsSourceTelegram, db.ClientsSourceReferall, db.ClientsSourceOrganic, db.ClientsSourceRepeatClient}
	jobs := []db.EmploymentStatus{db.EmploymentStatusFullTime, db.EmploymentStatusPartTime, db.EmploymentStatusSelfEmployed}
	methods := []db.BrokerPaymentMethod{db.BrokerPaymentMethodCash, db.BrokerPaymentMethodZelleGmotion, db.BrokerPaymentMethodZellePersonal}
	for i, s := range scenarios {
		status := db.ClientStatusActive
		if s.paid {
			status = db.ClientStatusClosed
		}
		if s.cancelled {
			status = db.ClientStatusNoDeal
		}
		dob := time.Date(1985+i%10, time.Month(1+i%12), 12, 0, 0, 0, 0, time.UTC)
		input := service.CreateClientInput{FirstName: s.first, LastName: s.last, Email: ptr(fmt.Sprintf("%s-%02d@example.test", marker, i)), Phone: ptr(fmt.Sprintf("202-555-%04d", 100+i)), Status: status,
			Source: db.NullClientsSource{ClientsSource: sources[i%len(sources)], Valid: true}, Notes: ptr("DEMO ONLY. Scenario: " + s.key),
			DateOfBirth: &dob, Address: ptr(fmt.Sprintf("%d Example Avenue", 100+i)), City: ptr("Los Angeles"), State: ptr("CA"), Zip: ptr("90001"), Ssn: ptr("000-12-3456"),
			HousingStatus: ptr([]string{"own", "rent", "family"}[i%3]), HousingDuration: ptr("3 years"), HousingMonthlyPayment: ptr(1800.0),
			EmployerName: ptr("Example Demo Company"), EmployerAddress: ptr("200 Sample Street"), EmployerCity: ptr("Los Angeles"), EmployerState: ptr("CA"), EmployerZip: ptr("90002"), EmployerPhone: ptr("202-555-0198"), Occupation: ptr("Demo analyst"),
			EmploymentStatus: db.NullEmploymentStatus{EmploymentStatus: jobs[i%len(jobs)], Valid: true}, EmploymentDuration: ptr("4 years"), GrossMonthlyIncome: ptr(6500.0 + float64(i)*200), OtherMonthlyIncome: ptr(250.0), OtherIncomeSource: ptr("Fictional side income")}
		if s.key == "new-incomplete" {
			input = service.CreateClientInput{FirstName: s.first, LastName: s.last, Email: input.Email, Status: db.ClientStatusActive, Notes: ptr("DEMO ONLY. Incomplete email-only lead")}
		}
		c, err := clients.Create(ctx, input)
		if err != nil {
			return err
		}
		createdClients = append(createdClients, c)
		date := time.Now().AddDate(0, 0, -s.age)
		if s.key == "lease-ending-soon" {
			date = time.Now().AddDate(0, -35, 0)
		}
		if s.key == "lease-ended" {
			date = time.Now().AddDate(0, -26, 0)
		}
		in := service.CreateDealInput{ClientID: id(c.ID), BrokerID: ptr(id(broker.ID)), DealerID: ptr(id(dealer.ID)), DealDate: &date,
			Make: ptr(s.make), Model: ptr(s.model), Vin: ptr(fmt.Sprintf("DEMO%013d", i+1)), TrimLevel: ptr("Demo trim"),
			BrokerFee: ptr(3000.0 + float64(i)*100), CollectAmount: ptr(500.0), DocFee: ptr(350.0), BrokerPercent: ptr(25.0),
			DownPayment: ptr(2500.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCard, Valid: true}, MonthlyPayment: ptr(550.0),
			IsCashDeal: s.cash, Notes: ptr("[" + marker + ":" + s.key + "] Fictional sample for local testing.")}
		if s.vinless {
			in.Vin = nil
			in.BrokerFee = nil
			in.CollectAmount = nil
			in.DocFee = nil
			in.BrokerPercent = nil
			in.MonthlyPayment = nil
			in.DownPayment = nil
		}
		if s.cash {
			in.DownPayment = ptr(32000.0)
			in.DownPaymentMethod = db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true}
			in.MonthlyPayment = nil
		}
		if s.months > 0 {
			in.LeaseTermMonths = ptr(int32(s.months))
		}
		if s.direct {
			in.BrokerID = nil
			in.BrokerPercent = nil
			in.BrokerCommission = ptr(0.0)
		}
		if s.key == "zero-fee-direct" {
			in.BrokerFee = ptr(0.0)
			in.CollectAmount = ptr(0.0)
			in.DocFee = ptr(0.0)
		}
		if s.key == "commission-unpaid" {
			in.BrokerPercent = nil
			in.BrokerCommission = ptr(900.0)
		}
		d, err := deals.Create(ctx, in)
		if err != nil {
			return fmt.Errorf("%s: %w", s.key, err)
		}
		statusDeal := db.DealStatusPending
		if s.paid {
			statusDeal = db.DealStatusPaid
		}
		if s.cancelled {
			statusDeal = db.DealStatusCancelled
		}
		d, err = deals.UpdateStatus(ctx, id(d.ID), statusDeal)
		if err != nil {
			return err
		}
		if s.paid && !s.commissionPending {
			if _, err = deals.UpdateBrokerCommissionStatus(ctx, id(d.ID), db.PaymentStatusPaid, db.NullBrokerPaymentMethod{BrokerPaymentMethod: methods[i%len(methods)], Valid: !s.direct}); err != nil {
				return err
			}
		}
		if s.key == "overdue-direct" {
			if _, err = deals.MarkPaymentReminderEmailSent(ctx, id(d.ID)); err != nil {
				return err
			}
		}
		createdDeals = append(createdDeals, d)
		if _, err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{EntityType: "deals", EntityID: d.ID, Actor: "local-demo-seed", Action: "create", ChangedFields: []byte(fmt.Sprintf(`{"scenario":%q,"demo":true}`, s.key))}); err != nil {
			return err
		}
		fmt.Printf("%s: %s %s — %s\n", d.DealNumber.String, s.first, s.last, s.key)
	}
	// Reproduce the cosigner relationship and cancellation of their redundant VIN-less intake.
	if err = q.AddDealCosigner(ctx, db.AddDealCosignerParams{DealID: createdDeals[9].ID, ClientID: createdClients[8].ID}); err != nil {
		return err
	}
	if _, err = q.CancelVinlessDealsForCosigner(ctx, db.CancelVinlessDealsForCosignerParams{ClientID: createdClients[8].ID, TargetDealID: createdDeals[9].ID, CancellationNote: "Demo: attached as cosigner to " + createdDeals[9].DealNumber.String}); err != nil {
		return err
	}
	// A repeat client with a second vehicle, without duplicating their client record.
	if _, err = deals.Create(ctx, service.CreateDealInput{ClientID: id(createdClients[0].ID), Make: ptr("Toyota"), Model: ptr("Corolla"), Vin: ptr("DEMO0000000000099"), BrokerFee: ptr(2200.0), Notes: ptr("[" + marker + ":repeat-client] Second vehicle for the same client")}); err != nil {
		return err
	}
	for i, status := range []db.ReferralStatus{db.ReferralStatusNone, db.ReferralStatusPending, db.ReferralStatusPaid} {
		for _, direct := range []bool{false, true} {
			in := service.CreateReferralInput{FirstName: "Demo", LastName: fmt.Sprintf("Referral%d%t", i, direct), Phone: ptr(fmt.Sprintf("202-555-%04d", 160+i)), Status: db.NullReferralStatus{ReferralStatus: status, Valid: true}, Notes: ptr("Fictional referral scenario"), Amount: ptr(300.0 + float64(i)*100)}
			if !direct {
				in.DealID = ptr(id(createdDeals[i].ID))
			}
			if status == db.ReferralStatusNone {
				in.Amount = nil
			}
			if _, err = refs.Create(ctx, in); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func seedDocuments(ctx context.Context, pool *pgxpool.Pool, documents *service.DocumentService) error {
	var clientID, dealID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT d.client_id,d.id FROM deals d JOIN clients c ON c.id=d.client_id WHERE c.email=$1 AND d.notes LIKE $2`, marker+"-00@example.test", "%:"+scenarios[0].key+"]%").Scan(&clientID, &dealID); err != nil {
		return err
	}
	types := []db.DocumentType{db.DocumentTypeCreditAppPdf, db.DocumentTypeDriverLicense, db.DocumentTypeInsurance, db.DocumentTypeBrokerAgreement, db.DocumentTypeDealerDocs, db.DocumentTypeCheckCopy, db.DocumentTypeOther}
	for i, t := range types {
		filename := "demo-" + string(t) + ".txt"
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM documents WHERE file_name=$1 AND (client_id=$2 OR deal_id=$3))`, filename, clientID, dealID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		body := "DEMO ONLY - NOT A VALID DOCUMENT\nFictional local CRM sample: " + string(t) + "\nContains no real personal information.\n"
		ownerClient, ownerDeal := uuid.NullUUID{}, uuid.NullUUID{}
		if i < 3 {
			ownerClient = uuid.NullUUID{UUID: clientID, Valid: true}
		} else {
			ownerDeal = uuid.NullUUID{UUID: dealID, Valid: true}
		}
		if _, err := documents.Upload(ctx, strings.NewReader(body), filename, int64(len(body)), "text/plain", ownerClient, ownerDeal, t); err != nil {
			return err
		}
	}
	return nil
}
