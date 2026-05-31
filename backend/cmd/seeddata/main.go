// seeddata populates the database with realistic test data for local development.
// Usage: cd backend && go run ./cmd/seeddata
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

func ptr[T any](v T) *T { return &v }

func daysAgo(n int) *time.Time {
	t := time.Now().AddDate(0, 0, -n)
	return &t
}

func main() {
	if os.Getenv("DATABASE_URL") == "" {
		_ = godotenv.Load()
		_ = godotenv.Load("../.env")
		_ = godotenv.Load("../../.env")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	encHex := os.Getenv("ENCRYPTION_KEY")
	if encHex == "" {
		log.Fatal("ENCRYPTION_KEY not set")
	}
	encKey, err := hex.DecodeString(encHex)
	if err != nil {
		log.Fatalf("invalid ENCRYPTION_KEY: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	brokerSvc := service.NewBroker(queries)
	dealerSvc := service.NewDealer(queries)
	clientSvc := service.NewClient(queries, encKey)
	dealSvc := service.NewDeal(queries, pool)
	referralSvc := service.NewReferral(queries)

	// ── Brokers ──────────────────────────────────────────────────────────────

	fmt.Println("Creating brokers...")
	brokerInputs := []service.CreateBrokerInput{
		{FirstName: "Marcus", LastName: "Rodriguez", Phone: ptr("310-555-0142"), Email: ptr("marcus.r@brokermail.com")},
		{FirstName: "Alicia", LastName: "Chen", Phone: ptr("818-555-0278"), Email: ptr("alicia.chen@premiumauto.com")},
		{FirstName: "Devon", LastName: "Williams", Phone: ptr("323-555-0391")},
	}
	brokers := make([]db.Broker, 0, len(brokerInputs))
	for _, in := range brokerInputs {
		b, err := brokerSvc.Create(ctx, in)
		if err != nil {
			log.Fatalf("create broker %s: %v", in.FirstName, err)
		}
		brokers = append(brokers, b)
		fmt.Printf("  broker: %s %s\n", b.FirstName, b.LastName)
	}

	// ── Dealers ───────────────────────────────────────────────────────────────

	fmt.Println("Creating dealers...")
	dealerInputs := []service.CreateDealerInput{
		{Name: "Pacific Motors of LA", ContactPerson: ptr("Tom Hsieh"), Phone: ptr("213-555-0801"), Address: ptr("4200 Wilshire Blvd, Los Angeles, CA 90010"), ContactPersonEmail: ptr("tom@pacificmotors.com")},
		{Name: "Bay Area Premium Auto", ContactPerson: ptr("Sandra Liu"), Phone: ptr("415-555-0562"), ContactPersonEmail: ptr("sliu@bapremium.com")},
		{Name: "Sunset Honda", ContactPerson: ptr("Ray Morales"), Phone: ptr("310-555-0934"), Address: ptr("8888 Sunset Blvd, West Hollywood, CA 90069")},
	}
	dealers := make([]db.Dealer, 0, len(dealerInputs))
	for _, in := range dealerInputs {
		d, err := dealerSvc.Create(ctx, in)
		if err != nil {
			log.Fatalf("create dealer %s: %v", in.Name, err)
		}
		dealers = append(dealers, d)
		fmt.Printf("  dealer: %s\n", d.Name)
	}

	// ── Clients ───────────────────────────────────────────────────────────────

	fmt.Println("Creating clients...")
	clientInputs := []service.CreateClientInput{
		{FirstName: "James", LastName: "Thompson", Phone: ptr("310-555-1001"), Email: ptr("james.thompson@gmail.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceInstagram, Valid: true}},
		{FirstName: "Sofia", LastName: "Ramirez", Phone: ptr("818-555-2002"), Email: ptr("sofia.ramirez@icloud.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceReferall, Valid: true}},
		{FirstName: "Alex", LastName: "Kim", Phone: ptr("213-555-3003"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceTelegram, Valid: true}},
		{FirstName: "Nina", LastName: "Patel", Phone: ptr("626-555-4004"), Email: ptr("nina.patel@yahoo.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceOrganic, Valid: true}},
		{FirstName: "Michael", LastName: "Johnson", Phone: ptr("323-555-5005"), Email: ptr("mjohnson@hotmail.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceInstagram, Valid: true}},
		{FirstName: "Carmen", LastName: "Torres", Phone: ptr("310-555-6006"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceReferall, Valid: true}},
		{FirstName: "Ryan", LastName: "Sullivan", Phone: ptr("424-555-7007"), Email: ptr("ryan.s@gmail.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceOrganic, Valid: true}},
		{FirstName: "Priya", LastName: "Sharma", Phone: ptr("818-555-8008"), Email: ptr("priya.sharma@gmail.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceTelegram, Valid: true}},
		{FirstName: "Carlos", LastName: "Martinez", Phone: ptr("714-555-9009"), Status: db.ClientStatusClosed, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceRepeatClient, Valid: true}},
		{FirstName: "Emily", LastName: "Nguyen", Phone: ptr("562-555-0010"), Email: ptr("emily.nguyen@icloud.com"), Status: db.ClientStatusActive, Source: db.NullClientsSource{ClientsSource: db.ClientsSourceInstagram, Valid: true}},
	}
	clients := make([]db.Client, 0, len(clientInputs))
	for _, in := range clientInputs {
		c, err := clientSvc.Create(ctx, in)
		if err != nil {
			log.Fatalf("create client %s: %v", in.FirstName, err)
		}
		clients = append(clients, c)
		fmt.Printf("  client: %s %s\n", c.FirstName, c.LastName)
	}

	// ── Deals ─────────────────────────────────────────────────────────────────

	fmt.Println("Creating deals...")

	brokerID := func(i int) *uuid.UUID {
		id, _ := uuid.FromBytes(brokers[i].ID.Bytes[:])
		return &id
	}
	dealerID := func(i int) *uuid.UUID {
		id, _ := uuid.FromBytes(dealers[i].ID.Bytes[:])
		return &id
	}
	clientID := func(i int) uuid.UUID {
		id, _ := uuid.FromBytes(clients[i].ID.Bytes[:])
		return id
	}

	paid := db.NullDealStatus{DealStatus: db.DealStatusPaid, Valid: true}
	pending := db.NullDealStatus{DealStatus: db.DealStatusPending, Valid: true}
	paidStatus := db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPaid, Valid: true}
	pendingStatus := db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPending, Valid: true}

	dealInputs := []service.CreateDealInput{
		// 1 — paid deal, Toyota Camry, broker 0, dealer 0
		{
			ClientID: clientID(0), BrokerID: brokerID(0), DealerID: dealerID(0),
			DealDate: daysAgo(62), Make: ptr("Toyota"), Model: ptr("Camry"), TrimLevel: ptr("XSE"),
			Vin: ptr("4T1G11AK5PU123456"), BrokerFee: ptr(3200.0), CollectAmount: ptr(800.0),
			DocFee: ptr(400.0), BrokerPercent: ptr(30.0),
			DealStatus: paid, BrokerFeeStatus: paidStatus, BrokerCommissionStatus: paidStatus,
			DownPayment: ptr(2500.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(520.0), Notes: ptr("Smooth deal, client very happy"),
		},
		// 2 — paid deal, Honda Accord, broker 1, dealer 2
		{
			ClientID: clientID(1), BrokerID: brokerID(1), DealerID: dealerID(2),
			DealDate: daysAgo(55), Make: ptr("Honda"), Model: ptr("Accord"), TrimLevel: ptr("Sport 2.0T"),
			Vin: ptr("1HGCV2F34PA234567"), BrokerFee: ptr(2800.0), CollectAmount: ptr(600.0),
			DocFee: ptr(350.0), BrokerPercent: ptr(25.0),
			DealStatus: paid, BrokerFeeStatus: paidStatus, BrokerCommissionStatus: paidStatus,
			DownPayment: ptr(3000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCard, Valid: true},
			MonthlyPayment: ptr(480.0),
		},
		// 3 — paid deal, BMW 3 Series, broker 0, dealer 1
		{
			ClientID: clientID(2), BrokerID: brokerID(0), DealerID: dealerID(1),
			DealDate: daysAgo(47), Make: ptr("BMW"), Model: ptr("3 Series"), TrimLevel: ptr("330i xDrive"),
			Vin: ptr("WBA5R7C5XKF345678"), BrokerFee: ptr(5500.0), CollectAmount: ptr(1200.0),
			DocFee: ptr(495.0), BrokerPercent: ptr(28.0),
			DealStatus: paid, BrokerFeeStatus: paidStatus, BrokerCommissionStatus: paidStatus,
			DownPayment: ptr(5000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(820.0),
		},
		// 4 — pending deal, Mercedes C-Class, broker 2, dealer 0
		{
			ClientID: clientID(3), BrokerID: brokerID(2), DealerID: dealerID(0),
			DealDate: daysAgo(38), Make: ptr("Mercedes-Benz"), Model: ptr("C-Class"), TrimLevel: ptr("C300 4MATIC"),
			Vin: ptr("55SWF4KB5LU456789"), BrokerFee: ptr(4800.0), CollectAmount: ptr(1000.0),
			DocFee: ptr(495.0), BrokerPercent: ptr(30.0),
			DealStatus: pending, BrokerFeeStatus: pendingStatus, BrokerCommissionStatus: pendingStatus,
			DownPayment: ptr(4500.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(750.0),
		},
		// 5 — pending deal, Hyundai Sonata, broker 1, dealer 2
		{
			ClientID: clientID(4), BrokerID: brokerID(1), DealerID: dealerID(2),
			DealDate: daysAgo(29), Make: ptr("Hyundai"), Model: ptr("Sonata"), TrimLevel: ptr("Limited"),
			Vin: ptr("5NPE34AF4KH567890"), BrokerFee: ptr(2200.0), CollectAmount: ptr(500.0),
			DocFee: ptr(300.0), BrokerPercent: ptr(22.0),
			DealStatus: pending, BrokerFeeStatus: pendingStatus, BrokerCommissionStatus: pendingStatus,
			DownPayment: ptr(2000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCard, Valid: true},
			MonthlyPayment: ptr(390.0),
		},
		// 6 — paid deal, Lexus RX, no broker, dealer 1
		{
			ClientID: clientID(5), DealerID: dealerID(1),
			DealDate: daysAgo(25), Make: ptr("Lexus"), Model: ptr("RX"), TrimLevel: ptr("RX350 F Sport"),
			Vin: ptr("2T2BZMCA2KC678901"), BrokerFee: ptr(4200.0), CollectAmount: ptr(900.0),
			DocFee: ptr(450.0),
			DealStatus: paid, BrokerFeeStatus: paidStatus,
			DownPayment: ptr(6000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(690.0), Notes: ptr("Direct client, no broker involved"),
		},
		// 7 — pending deal, Kia Telluride, broker 0, dealer 0
		{
			ClientID: clientID(6), BrokerID: brokerID(0), DealerID: dealerID(0),
			DealDate: daysAgo(18), Make: ptr("Kia"), Model: ptr("Telluride"), TrimLevel: ptr("SX-P AWD"),
			Vin: ptr("5XYP64HC5MG789012"), BrokerFee: ptr(3600.0), CollectAmount: ptr(750.0),
			DocFee: ptr(395.0), BrokerPercent: ptr(27.0),
			DealStatus: pending, BrokerFeeStatus: pendingStatus, BrokerCommissionStatus: pendingStatus,
			DownPayment: ptr(3500.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(610.0),
		},
		// 8 — pending deal, Audi Q5, broker 2, dealer 1
		{
			ClientID: clientID(7), BrokerID: brokerID(2), DealerID: dealerID(1),
			DealDate: daysAgo(12), Make: ptr("Audi"), Model: ptr("Q5"), TrimLevel: ptr("Premium Plus 45 TFSI"),
			Vin: ptr("WA1BNAFY4J2890123"), BrokerFee: ptr(5100.0), CollectAmount: ptr(1100.0),
			DocFee: ptr(495.0), BrokerPercent: ptr(29.0),
			DealStatus: pending, BrokerFeeStatus: pendingStatus, BrokerCommissionStatus: pendingStatus,
			DownPayment: ptr(5500.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCard, Valid: true},
			MonthlyPayment: ptr(880.0),
		},
		// 9 — paid deal (repeat client), Chevrolet Equinox, broker 1, dealer 2
		{
			ClientID: clientID(8), BrokerID: brokerID(1), DealerID: dealerID(2),
			DealDate: daysAgo(8), Make: ptr("Chevrolet"), Model: ptr("Equinox"), TrimLevel: ptr("Premier AWD"),
			Vin: ptr("3GNAXKEV3KS901234"), BrokerFee: ptr(2600.0), CollectAmount: ptr(550.0),
			DocFee: ptr(325.0), BrokerPercent: ptr(24.0),
			DealStatus: paid, BrokerFeeStatus: paidStatus, BrokerCommissionStatus: paidStatus,
			DownPayment: ptr(2000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCash, Valid: true},
			MonthlyPayment: ptr(440.0), Notes: ptr("Repeat client — 2nd deal"),
		},
		// 10 — pending deal (brand new, 3 days ago), Tesla Model 3, broker 0, dealer 0
		{
			ClientID: clientID(9), BrokerID: brokerID(0), DealerID: dealerID(0),
			DealDate: daysAgo(3), Make: ptr("Tesla"), Model: ptr("Model 3"), TrimLevel: ptr("Long Range AWD"),
			Vin: ptr("5YJ3E1EA4KF012345"), BrokerFee: ptr(4500.0), CollectAmount: ptr(950.0),
			DocFee: ptr(450.0), BrokerPercent: ptr(26.0),
			DealStatus: pending, BrokerFeeStatus: pendingStatus, BrokerCommissionStatus: pendingStatus,
			DownPayment: ptr(4000.0), DownPaymentMethod: db.NullDownPaymentMethod{DownPaymentMethod: db.DownPaymentMethodCard, Valid: true},
			MonthlyPayment: ptr(720.0),
		},
	}

	deals := make([]db.Deal, 0, len(dealInputs))
	for i, in := range dealInputs {
		d, err := dealSvc.Create(ctx, in)
		if err != nil {
			log.Fatalf("create deal %d: %v", i+1, err)
		}
		deals = append(deals, d)
		fmt.Printf("  deal %s: %s %s (%s)\n", d.DealNumber.String, d.Make.String, d.Model.String, d.DealStatus.DealStatus)
	}

	// ── Referrals (on a few deals) ────────────────────────────────────────────

	fmt.Println("Creating referrals...")

	type referralSeed struct {
		dealIdx   int
		firstName string
		lastName  string
		phone     string
		amount    float64
		status    db.ReferralStatus
	}

	referralSeeds := []referralSeed{
		{dealIdx: 0, firstName: "Miguel", lastName: "Reyes", phone: "310-555-7701", amount: 300.0, status: db.ReferralStatusPaid},
		{dealIdx: 2, firstName: "Tina", lastName: "Brooks", phone: "818-555-4402", amount: 500.0, status: db.ReferralStatusPaid},
		{dealIdx: 3, firstName: "David", lastName: "Park", phone: "323-555-5503", amount: 400.0, status: db.ReferralStatusPending},
		{dealIdx: 6, firstName: "Laura", lastName: "Castillo", phone: "424-555-6604", amount: 350.0, status: db.ReferralStatusPending},
	}

	for _, r := range referralSeeds {
		dealUUID, _ := uuid.FromBytes(deals[r.dealIdx].ID.Bytes[:])
		ref, err := referralSvc.Create(ctx, service.CreateReferralInput{
			DealID:    &dealUUID,
			FirstName: r.firstName,
			LastName:  r.lastName,
			Phone:     ptr(r.phone),
			Amount:    ptr(r.amount),
			Status:    db.NullReferralStatus{ReferralStatus: r.status, Valid: true},
		})
		if err != nil {
			log.Fatalf("create referral for deal %d: %v", r.dealIdx+1, err)
		}
		fmt.Printf("  referral: %s %s ($%.0f) on deal %s\n",
			ref.FirstName, ref.LastName, r.amount, deals[r.dealIdx].DealNumber.String)
	}

	fmt.Println("\nDone! Seeded:")
	fmt.Printf("  %d brokers, %d dealers, %d clients, %d deals, %d referrals\n",
		len(brokers), len(dealers), len(clients), len(deals), len(referralSeeds))
}
