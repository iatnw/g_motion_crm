package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func (s *DealService) WithTx(tx pgx.Tx) *DealService {
	return &DealService{q: s.q.WithTx(tx), pool: s.pool}
}

type DealService struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewDeal(q *db.Queries, pool *pgxpool.Pool) *DealService {
	return &DealService{q: q, pool: pool}
}

func (s *DealService) PreviewNextDealNumber(ctx context.Context) (string, error) {
	var seq int64
	err := s.pool.QueryRow(ctx,
		`SELECT CASE WHEN is_called THEN last_value + 1 ELSE last_value END FROM deal_number_seq`,
	).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatDealNumber(seq), nil
}

type CreateDealInput struct {
	ClientID               uuid.UUID
	BrokerID               *uuid.UUID
	DealerID               *uuid.UUID
	DealDate               *time.Time
	DealNumber             *string
	Vin                    *string
	Make                   *string
	Model                  *string
	TrimLevel              *string
	DownPayment            *float64
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         *float64
	BrokerFee              *float64
	BrokerPercent          *float64
	BrokerCommission       *float64
	BrokerCommissionStatus db.NullPaymentStatus
	BrokerFeeStatus        db.NullPaymentStatus
	DealStatus             db.NullDealStatus
	CollectAmount          *float64
	DocFee                 *float64
	BrokerPaymentMethod    db.NullBrokerPaymentMethod
	Notes                  *string
	IsCashDeal             bool
	LeaseTermMonths        *int32
}

func (s *DealService) Create(ctx context.Context, in CreateDealInput) (db.Deal, error) {
	if !in.DealStatus.Valid {
		in.DealStatus = db.NullDealStatus{DealStatus: db.DealStatusPending, Valid: true}
	}
	if !in.BrokerFeeStatus.Valid {
		in.BrokerFeeStatus = db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPending, Valid: true}
	}
	if !in.BrokerCommissionStatus.Valid {
		in.BrokerCommissionStatus = db.NullPaymentStatus{PaymentStatus: db.PaymentStatusPending, Valid: true}
	}
	if in.DealDate == nil {
		now := time.Now()
		in.DealDate = &now
	}

	if err := normalizeBrokerCommission(in.BrokerFee, in.DocFee, in.CollectAmount, &in.BrokerPercent, &in.BrokerCommission); err != nil {
		return db.Deal{}, err
	}

	seq, err := s.q.NextDealNumber(ctx)
	if err != nil {
		return db.Deal{}, err
	}
	num := formatDealNumber(seq)
	in.DealNumber = &num

	deal, err := s.q.CreateDeal(ctx, db.CreateDealParams{
		ClientID:               toUUID(in.ClientID),
		BrokerID:               toNullUUID(in.BrokerID),
		DealerID:               toNullUUID(in.DealerID),
		DealDate:               toDate(in.DealDate),
		DealNumber:             toText(in.DealNumber),
		Vin:                    toText(in.Vin),
		Make:                   toText(in.Make),
		Model:                  toText(in.Model),
		TrimLevel:              toText(in.TrimLevel),
		DownPayment:            toNumeric(in.DownPayment),
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         toNumeric(in.MonthlyPayment),
		BrokerFee:              toNumeric(in.BrokerFee),
		BrokerPercent:          toNumeric(in.BrokerPercent),
		BrokerCommission:       toNumeric(in.BrokerCommission),
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		BrokerFeeStatus:        in.BrokerFeeStatus,
		DealStatus:             in.DealStatus,
		CollectAmount:          toNumeric(in.CollectAmount),
		DocFee:                 toNumeric(in.DocFee),
		BrokerPaymentMethod:    in.BrokerPaymentMethod,
		Notes:                  toText(in.Notes),
		IsCashDeal:             in.IsCashDeal,
		LeaseTermMonths:        toInt32(in.LeaseTermMonths),
	})
	if err != nil {
		return db.Deal{}, dealWriteError(err)
	}
	return deal, nil
}

func formatDealNumber(seq int64) string {
	return fmt.Sprintf("GM-%04d", seq)
}

func (s *DealService) Get(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.GetDeal(ctx, toUUID(id))
}

func (s *DealService) GetShow(ctx context.Context, id uuid.UUID) (db.GetDealShowRow, error) {
	return s.q.GetDealShow(ctx, toUUID(id))
}

type DealListCounts struct {
	All       int64
	New       int64
	Pending   int64
	Paid      int64
	Cancelled int64
}

func (s *DealService) ListFiltered(ctx context.Context, status, search, vinFilter, leaseFilter string, leaseWithin *int32, limit, offset int32) ([]db.Deal, DealListCounts, error) {
	var leaseFilterParam pgtype.Text
	if leaseFilter != "" {
		leaseFilterParam = pgtype.Text{String: leaseFilter, Valid: true}
	}
	rows, err := s.q.ListDealsFiltered(ctx, db.ListDealsFilteredParams{
		Status:      status,
		Search:      search,
		VinFilter:   vinFilter,
		LeaseFilter: leaseFilterParam,
		LeaseWithin: toInt32(leaseWithin),
		Lim:         limit,
		Off:         offset,
	})
	if err != nil {
		return nil, DealListCounts{}, err
	}

	var counts DealListCounts
	deals := make([]db.Deal, 0, len(rows))
	for i, row := range rows {
		if i == 0 {
			counts = DealListCounts{
				All:       row.AllCount,
				New:       row.NewCount,
				Pending:   row.PendingCount,
				Paid:      row.PaidCount,
				Cancelled: row.CancelledCount,
			}
		}
		if !row.ID.Valid {
			continue
		}
		deals = append(deals, dealFromFilteredRow(row))
	}

	return deals, counts, nil
}

func dealFromFilteredRow(row db.ListDealsFilteredRow) db.Deal {
	return db.Deal{
		ID:                         row.ID,
		ClientID:                   row.ClientID,
		BrokerID:                   row.BrokerID,
		DealerID:                   row.DealerID,
		DealDate:                   row.DealDate,
		DealNumber:                 row.DealNumber,
		Vin:                        row.Vin,
		Make:                       row.Make,
		Model:                      row.Model,
		TrimLevel:                  row.TrimLevel,
		DownPayment:                row.DownPayment,
		DownPaymentMethod:          row.DownPaymentMethod,
		MonthlyPayment:             row.MonthlyPayment,
		BrokerFee:                  row.BrokerFee,
		BrokerPercent:              row.BrokerPercent,
		BrokerCommission:           row.BrokerCommission,
		BrokerCommissionStatus:     row.BrokerCommissionStatus,
		BrokerCommissionPaidDate:   row.BrokerCommissionPaidDate,
		YourNet:                    row.YourNet,
		BrokerFeeStatus:            row.BrokerFeeStatus,
		BrokerFeePaidDate:          row.BrokerFeePaidDate,
		CreatedAt:                  row.CreatedAt,
		UpdatedAt:                  row.UpdatedAt,
		DealStatus:                 row.DealStatus,
		CollectAmount:              row.CollectAmount,
		DocFee:                     row.DocFee,
		BrokerPaymentMethod:        row.BrokerPaymentMethod,
		Notes:                      row.Notes,
		DealStatusPendingSince:     row.DealStatusPendingSince,
		PaymentReminderEmailSentAt: row.PaymentReminderEmailSentAt,
		IsCashDeal:                 row.IsCashDeal.Bool,
		LeaseTermMonths:            row.LeaseTermMonths,
	}
}

func (s *DealService) ListByClient(ctx context.Context, id uuid.UUID, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDealsByClient(ctx, db.ListDealsByClientParams{
		ClientID: toUUID(id),
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *DealService) ListByCosigner(ctx context.Context, id uuid.UUID, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDealsByCosigner(ctx, db.ListDealsByCosignerParams{
		ClientID: toUUID(id),
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *DealService) ListByBroker(ctx context.Context, id uuid.UUID, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDealsByBroker(ctx, db.ListDealsByBrokerParams{
		BrokerID: toNullUUID(&id),
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *DealService) ListByDealer(ctx context.Context, id uuid.UUID, limit, offset int32) ([]db.Deal, error) {
	return s.q.ListDealsByDealer(ctx, db.ListDealsByDealerParams{
		DealerID: toNullUUID(&id),
		Limit:    limit,
		Offset:   offset,
	})
}

type UpdateDealInput struct {
	ID                     uuid.UUID
	ClientID               uuid.UUID
	BrokerID               *uuid.UUID
	DealerID               *uuid.UUID
	DealDate               *time.Time
	DealNumber             *string
	Vin                    *string
	Make                   *string
	Model                  *string
	TrimLevel              *string
	DownPayment            *float64
	DownPaymentMethod      db.NullDownPaymentMethod
	MonthlyPayment         *float64
	BrokerFee              *float64
	BrokerPercent          *float64
	BrokerCommission       *float64
	BrokerCommissionStatus db.NullPaymentStatus
	BrokerFeeStatus        db.NullPaymentStatus
	DealStatus             db.NullDealStatus
	CollectAmount          *float64
	DocFee                 *float64
	BrokerPaymentMethod    db.NullBrokerPaymentMethod
	Notes                  *string
	IsCashDeal             bool
	LeaseTermMonths        *int32
}

func (s *DealService) Update(ctx context.Context, in UpdateDealInput) (db.Deal, error) {
	if err := normalizeBrokerCommission(in.BrokerFee, in.DocFee, in.CollectAmount, &in.BrokerPercent, &in.BrokerCommission); err != nil {
		return db.Deal{}, err
	}

	deal, err := s.q.UpdateDeal(ctx, db.UpdateDealParams{
		ID:                     toUUID(in.ID),
		ClientID:               toUUID(in.ClientID),
		BrokerID:               toNullUUID(in.BrokerID),
		DealerID:               toNullUUID(in.DealerID),
		DealDate:               toDate(in.DealDate),
		DealNumber:             toText(in.DealNumber),
		Vin:                    toText(in.Vin),
		Make:                   toText(in.Make),
		Model:                  toText(in.Model),
		TrimLevel:              toText(in.TrimLevel),
		DownPayment:            toNumeric(in.DownPayment),
		DownPaymentMethod:      in.DownPaymentMethod,
		MonthlyPayment:         toNumeric(in.MonthlyPayment),
		BrokerFee:              toNumeric(in.BrokerFee),
		BrokerPercent:          toNumeric(in.BrokerPercent),
		BrokerCommission:       toNumeric(in.BrokerCommission),
		BrokerCommissionStatus: in.BrokerCommissionStatus,
		BrokerFeeStatus:        in.BrokerFeeStatus,
		DealStatus:             in.DealStatus,
		CollectAmount:          toNumeric(in.CollectAmount),
		DocFee:                 toNumeric(in.DocFee),
		BrokerPaymentMethod:    in.BrokerPaymentMethod,
		Notes:                  toText(in.Notes),
		IsCashDeal:             in.IsCashDeal,
		LeaseTermMonths:        toInt32(in.LeaseTermMonths),
	})
	return deal, dealWriteError(err)
}

func (s *DealService) UpdateBrokerFeeStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus) (db.Deal, error) {
	return s.q.UpdateDealBrokerFeeStatus(ctx, db.UpdateDealBrokerFeeStatusParams{
		ID:              toUUID(id),
		BrokerFeeStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
	})
}

func (s *DealService) UpdateBrokerCommissionStatus(ctx context.Context, id uuid.UUID, status db.PaymentStatus, method db.NullBrokerPaymentMethod) (db.Deal, error) {
	return s.q.UpdateDealBrokerCommissionStatus(ctx, db.UpdateDealBrokerCommissionStatusParams{
		ID:                     toUUID(id),
		BrokerCommissionStatus: db.NullPaymentStatus{PaymentStatus: status, Valid: true},
		BrokerPaymentMethod:    method,
	})
}

func (s *DealService) UpdateStatus(ctx context.Context, id uuid.UUID, status db.DealStatus) (db.Deal, error) {
	return s.q.UpdateDealStatus(ctx, db.UpdateDealStatusParams{
		ID:         toUUID(id),
		DealStatus: db.NullDealStatus{DealStatus: status, Valid: true},
	})
}

func (s *DealService) MarkPaymentReminderEmailSent(ctx context.Context, id uuid.UUID) (db.Deal, error) {
	return s.q.MarkDealPaymentReminderEmailSent(ctx, toUUID(id))
}

// AttachCosigner links a client to a deal and cancels that client's other VIN-less
// deals, which are no longer valid after the client is identified as a cosigner.
func (s *DealService) AttachCosigner(ctx context.Context, dealID, clientID uuid.UUID) ([]db.Deal, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin attach cosigner transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := s.q.WithTx(tx)
	target, err := q.GetDeal(ctx, toUUID(dealID))
	if err != nil {
		return nil, err
	}
	if target.DealStatus.Valid && target.DealStatus.DealStatus == db.DealStatusCancelled {
		return nil, fmt.Errorf("cannot attach a cosigner to a cancelled deal")
	}
	if target.ClientID.Valid && uuid.UUID(target.ClientID.Bytes) == clientID {
		return nil, fmt.Errorf("the primary client cannot also be a cosigner")
	}

	if err := q.AddDealCosigner(ctx, db.AddDealCosignerParams{
		DealID:   toUUID(dealID),
		ClientID: toUUID(clientID),
	}); err != nil {
		return nil, err
	}
	moveParams := db.MoveCosignerDocumentsToDealParams{
		ClientID:     toUUID(clientID),
		TargetDealID: toUUID(dealID),
	}
	if err := q.MoveCosignerDocumentsToDeal(ctx, moveParams); err != nil {
		return nil, err
	}
	if err := q.MovePendingCosignerDocumentImportsToDeal(ctx, db.MovePendingCosignerDocumentImportsToDealParams(moveParams)); err != nil {
		return nil, err
	}

	targetLabel := "this deal"
	if target.DealNumber.Valid && target.DealNumber.String != "" {
		targetLabel = "deal " + target.DealNumber.String
	}
	cancellationNote := "Cancelled automatically: client was attached as a cosigner to " + targetLabel + "."
	cancelled, err := q.CancelVinlessDealsForCosigner(ctx, db.CancelVinlessDealsForCosignerParams{
		ClientID:         toUUID(clientID),
		TargetDealID:     toUUID(dealID),
		CancellationNote: cancellationNote,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit attach cosigner transaction: %w", err)
	}
	return cancelled, nil
}

func (s *DealService) ListCosigners(ctx context.Context, dealID uuid.UUID) ([]db.Client, error) {
	return s.q.ListDealCosigners(ctx, toUUID(dealID))
}

func (s *DealService) RemoveCosigner(ctx context.Context, dealID, clientID uuid.UUID) error {
	deleted, err := s.q.DeleteDealCosigner(ctx, db.DeleteDealCosignerParams{
		DealID:   toUUID(dealID),
		ClientID: toUUID(clientID),
	})
	if err != nil {
		return err
	}
	if deleted == 0 {
		return fmt.Errorf("cosigner relationship not found")
	}
	return nil
}

func normalizeBrokerCommission(brokerFee *float64, docFee *float64, collectAmount *float64, brokerPercent **float64, brokerCommission **float64) error {
	fee := 0.0
	if brokerFee != nil {
		fee = *brokerFee
	}
	doc := 0.0
	if docFee != nil {
		doc = *docFee
	}
	collect := 0.0
	if collectAmount != nil {
		collect = *collectAmount
	}

	base := fee + collect - doc

	hasPercent := brokerPercent != nil && *brokerPercent != nil
	hasCommission := brokerCommission != nil && *brokerCommission != nil

	if hasPercent && !hasCommission {
		if base > 0 {
			commission := base * (**brokerPercent) / 100
			*brokerCommission = &commission
		}
	} else if hasCommission && !hasPercent {
		if base > 0 {
			percent := (**brokerCommission) / base * 100
			*brokerPercent = &percent
		}
	} else if hasPercent && hasCommission {
		// If percent is 0 but commission is set, recompute percent from commission
		if **brokerPercent == 0 && base > 0 {
			percent := (**brokerCommission) / base * 100
			*brokerPercent = &percent
		}
	}

	return nil
}
