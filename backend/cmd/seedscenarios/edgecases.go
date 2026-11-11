package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

// seedEdgeCases supplements workflow examples with boundary and pagination data.
// Its own marker allows upgrading an already-seeded local database safely.
func seedEdgeCases(ctx context.Context, pool *pgxpool.Pool, key []byte) error {
	const edgeMarker = "local-demo-edges-v1"
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, edgeMarker); err != nil {
		return err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM clients WHERE notes LIKE $1)`, "%["+edgeMarker+"]%").Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	q := db.New(tx)
	clients, deals := service.NewClient(q, key), service.NewDeal(q, pool)
	broker, err := service.NewBroker(q).Create(ctx, service.CreateBrokerInput{FirstName: "Elena", LastName: "García-López", Email: ptr("international-partner@example.test"), Notes: ptr("DEMO ONLY. International names and populated partner statistics.")})
	if err != nil {
		return err
	}
	dealer, err := service.NewDealer(q).Create(ctx, service.CreateDealerInput{Name: "Demo — West Coast Automotive & Commercial Vehicle Specialists", ContactPerson: ptr("Jean-Luc O’Neill"), Notes: ptr("DEMO ONLY. Long dealership name for responsive layouts.")})
	if err != nil {
		return err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	names := [][2]string{{"María José", "García-López"}, {"Jean-Luc", "O’Neill"}, {"李", "明"}, {"Zoë", "van der Westhuizen-Smith"}, {"Alexandria-Catherine", "Montgomery-Worthington"}, {"Sam", "Lee"}, {"Sam", "Lee"}}
	ages := []int{0, 1, 30, 31, 32, 39, 40, 41, 42, 365}
	vehicles := [][2]string{{"Toyota", "Camry"}, {"Honda", "Civic"}, {"Tesla", "Model Y"}, {"BMW", "X5"}, {"Lexus", "RX"}, {"Ford", "F-150"}, {"Hyundai", "Ioniq 5"}, {"Mercedes-Benz", "E-Class"}}
	statuses := []db.ClientStatus{db.ClientStatusActive, db.ClientStatusClosed, db.ClientStatusNoDeal}
	sources := []db.ClientsSource{db.ClientsSourceInstagram, db.ClientsSourceTelegram, db.ClientsSourceReferall, db.ClientsSourceOrganic, db.ClientsSourceRepeatClient}
	for i := 0; i < 205; i++ {
		first, last := fmt.Sprintf("Demo%03d", i+1), "Pagination"
		if i < len(names) {
			first, last = names[i][0], names[i][1]
		}
		note := fmt.Sprintf("[%s] DEMO ONLY. Pagination example %03d.", edgeMarker, i+1)
		in := service.CreateClientInput{FirstName: first, LastName: last, Email: ptr(fmt.Sprintf("edge-%03d@example.test", i)), Phone: ptr(fmt.Sprintf("202-555-%04d", 200+i)), Status: statuses[i%3], Source: db.NullClientsSource{ClientsSource: sources[i%len(sources)], Valid: true}, Notes: &note}
		switch i {
		case 0:
			in.Phone = nil
			in.Email = nil
		case 1:
			in.Email = nil
		case 2:
			in.Phone = nil
		case 3:
			in.Notes = ptr(note + " Long notes: client requested an accessible vehicle, multiple quotes, delivery coordination, and a detailed comparison before deciding.\nFollow-up: confirm insurance and preferred collection date.")
		}
		c, err := clients.Create(ctx, in)
		if err != nil {
			return err
		}
		age := (i * 7) % 180
		if i < len(ages) {
			age = ages[i]
		}
		date := today.AddDate(0, 0, -age)
		vehicle := vehicles[i%len(vehicles)]
		d := service.CreateDealInput{ClientID: id(c.ID), BrokerID: ptr(id(broker.ID)), DealerID: ptr(id(dealer.ID)), DealDate: &date, Make: ptr(vehicle[0]), Model: ptr(vehicle[1]), Vin: ptr(fmt.Sprintf("EDGE%013d", i+1)), BrokerFee: ptr(1200.25 + float64(i%20)*175.5), CollectAmount: ptr(250.75), DocFee: ptr(150.25), BrokerPercent: ptr(25.0), MonthlyPayment: ptr(399.99 + float64(i%10)*50), DownPayment: ptr(2000.0), Notes: ptr(note)}
		if i < len(ages) {
			d.Notes = ptr(fmt.Sprintf("[%s] DEMO ONLY. Aging boundary: %d days pending.", edgeMarker, age))
		}
		// Remaining lease months exercise every lease filter and the ended state.
		if i >= 10 && i < 16 {
			remaining := []int{0, 1, 3, 6, 12, 24}[i-10]
			date = today.AddDate(0, -(36 - remaining), 0)
			d.DealDate = &date
			d.LeaseTermMonths = ptr(int32(36))
			d.Notes = ptr(fmt.Sprintf("[%s] DEMO ONLY. Lease boundary: %d whole months remaining.", edgeMarker, remaining))
		}
		if i%11 == 0 && i >= 16 {
			d.IsCashDeal = true
			d.MonthlyPayment = nil
			d.DownPayment = ptr(32000.0)
		}
		if i%13 == 0 && i >= 16 {
			d.BrokerID = nil
			d.DealerID = nil
			d.BrokerPercent = nil
			d.BrokerCommission = ptr(0.0)
		}
		if i%7 == 0 && i >= 16 {
			d.BrokerPercent = nil
			d.BrokerCommission = ptr(550.55)
		}
		deal, err := deals.Create(ctx, d)
		if err != nil {
			return fmt.Errorf("edge scenario %d: %w", i, err)
		}
		if i >= 10 {
			status := db.DealStatusPaid
			if i >= 16 {
				status = []db.DealStatus{db.DealStatusPending, db.DealStatusPaid, db.DealStatusCancelled}[i%3]
			}
			if _, err = deals.UpdateStatus(ctx, id(deal.ID), status); err != nil {
				return err
			}
			if status == db.DealStatusPaid && d.BrokerID != nil && i%2 == 0 {
				method := []db.BrokerPaymentMethod{db.BrokerPaymentMethodCash, db.BrokerPaymentMethodZelleGmotion, db.BrokerPaymentMethodZellePersonal}[i%3]
				if _, err = deals.UpdateBrokerCommissionStatus(ctx, id(deal.ID), db.PaymentStatusPaid, db.NullBrokerPaymentMethod{BrokerPaymentMethod: method, Valid: true}); err != nil {
					return err
				}
			}
		}
	}
	// A client with no deals distinguishes an unconverted contact from a new deal.
	if _, err = clients.Create(ctx, service.CreateClientInput{FirstName: "Robin", LastName: "NoDealsYet", Status: db.ClientStatusActive, Notes: ptr("[" + edgeMarker + "] DEMO ONLY. Client with no deals, documents, or contact details.")}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// seedPreviewDocuments includes real file formats for download and image preview.
func seedPreviewDocuments(ctx context.Context, pool *pgxpool.Pool, documents *service.DocumentService) error {
	var clientID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM clients WHERE email=$1`, marker+"-00@example.test").Scan(&clientID); err != nil {
		return err
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 640, 400))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.RGBA{240, 245, 252, 255}), image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, 0, 640, 80), image.NewUniform(color.RGBA{30, 70, 130, 255}), image.Point{}, draw.Src)
	for _, r := range []image.Rectangle{image.Rect(35, 115, 175, 280), image.Rect(215, 120, 590, 140), image.Rect(215, 170, 520, 185), image.Rect(215, 215, 565, 230), image.Rect(35, 325, 590, 345)} {
		draw.Draw(canvas, r, image.NewUniform(color.RGBA{155, 180, 210, 255}), image.Point{}, draw.Src)
	}
	var picture bytes.Buffer
	if err := png.Encode(&picture, canvas); err != nil {
		return err
	}
	stream := "BT /F1 22 Tf 50 740 Td (DEMO ONLY - NOT A VALID DOCUMENT) Tj 0 -42 Td /F1 12 Tf (Fictional G Motion CRM sample. Contains no real personal information.) Tj ET"
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>", fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream)}
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, object := range objects {
		offsets = append(offsets, pdf.Len())
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	for _, file := range []struct {
		name, mime, body string
		kind             db.DocumentType
	}{
		{"DEMO-image-preview.png", "image/png", picture.String(), db.DocumentTypeOther},
		{"DEMO-credit-application-not-a-valid-document.pdf", "application/pdf", pdf.String(), db.DocumentTypeCreditAppPdf},
	} {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM documents WHERE client_id=$1 AND file_name=$2)`, clientID, file.name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := documents.Upload(ctx, strings.NewReader(file.body), file.name, int64(len(file.body)), file.mime, uuid.NullUUID{UUID: clientID, Valid: true}, uuid.NullUUID{}, file.kind); err != nil {
			return err
		}
	}
	return nil
}
