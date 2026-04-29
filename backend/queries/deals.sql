-- name: CreateDeal :one
INSERT INTO deals (
    client_id,
    broker_id,
    dealer_id,
    deal_date,
    deal_number,
    vin,
    make,
    model,
    trim_level,
    down_payment,
    down_payment_method,
    monthly_payment,
    broker_fee,
    broker_percent,
    broker_commission,
    broker_commission_status,
    broker_commission_paid_date,
    your_net,
    broker_fee_status,
    broker_fee_paid_date,
    deal_status,
    deal_status_pending_since,
    collect_amount,
    doc_fee,
    broker_payment_method,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21,
    CASE WHEN $21::deal_status = 'pending' THEN CURRENT_DATE ELSE NULL END,
    $22, $23, $24, $25
) RETURNING *;

-- name: GetDeal :one
SELECT * FROM deals
WHERE id = $1;

-- name: ListDeals :many
SELECT deals.* FROM deals
JOIN clients ON clients.id = deals.client_id
WHERE clients.status != 'no_deal'
ORDER BY deals.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateDeal :one
UPDATE deals
SET
    client_id                   = $2,
    broker_id                   = $3,
    dealer_id                   = $4,
    deal_date                   = $5,
    deal_number                 = $6,
    vin                         = $7,
    make                        = $8,
    model                       = $9,
    trim_level                  = $10,
    down_payment                = $11,
    down_payment_method         = $12,
    monthly_payment             = $13,
    broker_fee                  = $14,
    broker_percent              = $15,
    broker_commission           = $16,
    broker_commission_status    = $17,
    broker_commission_paid_date = $18,
    your_net                    = $19,
    broker_fee_status           = $20,
    broker_fee_paid_date        = $21,
    deal_status                 = $22,
    deal_status_pending_since   = CASE
        WHEN $22::deal_status = 'pending'
         AND (deal_status IS DISTINCT FROM 'pending' OR deal_status_pending_since IS NULL)
        THEN CURRENT_DATE
        WHEN $22::deal_status IS DISTINCT FROM 'pending'
        THEN NULL
        ELSE deal_status_pending_since
    END,
    collect_amount              = $23,
    doc_fee                     = $24,
    broker_payment_method       = $25,
    notes                       = $26,
    updated_at                  = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkDealPaymentReminderEmailSent :one
UPDATE deals
SET
    payment_reminder_email_sent_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListDealsByClient :many
SELECT * FROM deals
WHERE deals.client_id = $1
ORDER BY deals.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateDealBrokerFeeStatus :one
UPDATE deals
SET
    broker_fee_status    = $2,
    broker_fee_paid_date = CASE WHEN $2::payment_status = 'paid'
                           THEN CURRENT_DATE ELSE NULL END,
    updated_at           = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDealBrokerCommissionStatus :one
UPDATE deals
SET
    broker_commission_status    = $2,
    broker_commission_paid_date = CASE WHEN $2::payment_status = 'paid'
                                  THEN CURRENT_DATE ELSE NULL END,
    updated_at                  = NOW()
WHERE id = $1
RETURNING *;
