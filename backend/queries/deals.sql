-- name: NextDealNumber :one
SELECT nextval('deal_number_seq');

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
    CASE WHEN $21::deal_status = 'pending' THEN COALESCE($4::date, CURRENT_DATE) ELSE NULL END,
    $22, $23, $24, $25
) RETURNING *;

-- name: GetDeal :one
SELECT * FROM deals
WHERE id = $1;

-- name: GetDealShow :one
SELECT
    d.id,
    d.deal_date,
    (c.first_name || ' ' || c.last_name)::text AS client_name,
    d.make,
    d.model,
    d.trim_level,
    d.vin,
    dealers.name AS dealer_name,
    CASE WHEN brokers.id IS NOT NULL THEN (brokers.first_name || ' ' || brokers.last_name)::text ELSE ''::text END AS broker_name,
    d.broker_fee,
    d.collect_amount,
    d.doc_fee,
    COALESCE(referral_totals.referral_total, 0::numeric)::numeric AS referral_total,
    d.broker_percent,
    d.broker_commission,
    (
        COALESCE(d.broker_fee, 0)
        + COALESCE(d.collect_amount, 0)
        - COALESCE(d.broker_commission, 0)
        - COALESCE(referral_totals.referral_total, 0)
        - COALESCE(d.doc_fee, 0)
    )::numeric AS company_net,
    d.broker_fee_status,
    d.broker_commission_status,
    d.broker_payment_method,
    d.notes
FROM deals d
JOIN clients c ON c.id = d.client_id
LEFT JOIN dealers ON dealers.id = d.dealer_id
LEFT JOIN brokers ON brokers.id = d.broker_id
LEFT JOIN (
    SELECT
        deal_id,
        SUM(amount)::numeric AS referral_total
    FROM referrals
    GROUP BY deal_id
) referral_totals ON referral_totals.deal_id = d.id
WHERE d.id = $1;

-- name: ListDealsFiltered :many
WITH search_filtered AS (
    SELECT deals.* FROM deals
    JOIN clients ON clients.id = deals.client_id
    LEFT JOIN brokers ON brokers.id = deals.broker_id
    WHERE deals.deal_status IN ('pending', 'paid')
      AND (
        sqlc.arg(search)::text = ''
        OR deals.make ILIKE '%' || sqlc.arg(search) || '%'
        OR deals.model ILIKE '%' || sqlc.arg(search) || '%'
        OR deals.vin ILIKE '%' || sqlc.arg(search) || '%'
        OR deals.deal_number ILIKE '%' || sqlc.arg(search) || '%'
        OR clients.first_name ILIKE '%' || sqlc.arg(search) || '%'
        OR clients.last_name ILIKE '%' || sqlc.arg(search) || '%'
        OR (clients.first_name || ' ' || clients.last_name) ILIKE '%' || sqlc.arg(search) || '%'
        OR brokers.first_name ILIKE '%' || sqlc.arg(search) || '%'
        OR brokers.last_name ILIKE '%' || sqlc.arg(search) || '%'
        OR (brokers.first_name || ' ' || brokers.last_name) ILIKE '%' || sqlc.arg(search) || '%'
      )
),
counts AS (
    SELECT
        COUNT(*)::bigint AS all_count,
        COUNT(*) FILTER (WHERE deal_status = 'pending')::bigint AS pending_count,
        COUNT(*) FILTER (WHERE deal_status = 'paid')::bigint AS paid_count
    FROM search_filtered
),
page AS (
    SELECT *
    FROM search_filtered
    WHERE sqlc.arg(status)::text = '' OR deal_status::text = sqlc.arg(status)
    ORDER BY COALESCE(deal_date, created_at::date) DESC, created_at DESC
    LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off)
)
SELECT
    page.id,
    page.client_id,
    page.broker_id,
    page.dealer_id,
    page.deal_date,
    page.deal_number,
    page.vin,
    page.make,
    page.model,
    page.trim_level,
    page.down_payment,
    page.down_payment_method,
    page.monthly_payment,
    page.broker_fee,
    page.broker_percent,
    page.broker_commission,
    page.broker_commission_status,
    page.broker_commission_paid_date,
    page.your_net,
    page.broker_fee_status,
    page.broker_fee_paid_date,
    page.created_at,
    page.updated_at,
    page.deal_status,
    page.collect_amount,
    page.doc_fee,
    page.broker_payment_method,
    page.notes,
    page.deal_status_pending_since,
    page.payment_reminder_email_sent_at,
    counts.all_count,
    counts.pending_count,
    counts.paid_count
FROM counts
LEFT JOIN page ON true
ORDER BY COALESCE(page.deal_date, page.created_at::date) DESC, page.created_at DESC;

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

-- name: UpdateDealStatus :one
UPDATE deals
SET
    deal_status = $2,
    broker_fee_status = CASE
        WHEN $2::deal_status = 'paid' THEN 'paid'::payment_status
        WHEN $2::deal_status = 'pending' THEN 'pending'::payment_status
        ELSE broker_fee_status
    END,
    broker_fee_paid_date = CASE
        WHEN $2::deal_status = 'paid' THEN CURRENT_DATE
        WHEN $2::deal_status = 'pending' THEN NULL
        ELSE broker_fee_paid_date
    END,
    deal_status_pending_since = CASE
        WHEN $2::deal_status = 'pending'
         AND (deal_status IS DISTINCT FROM 'pending' OR deal_status_pending_since IS NULL)
        THEN CURRENT_DATE
        WHEN $2::deal_status IS DISTINCT FROM 'pending'
        THEN NULL
        ELSE deal_status_pending_since
    END,
    updated_at = NOW()
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

-- name: ListDealsByBroker :many
SELECT * FROM deals
WHERE deals.broker_id = $1
ORDER BY COALESCE(deals.deal_date, deals.created_at::date) DESC, deals.created_at DESC
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
    broker_payment_method       = $3,
    updated_at                  = NOW()
WHERE id = $1
RETURNING *;
