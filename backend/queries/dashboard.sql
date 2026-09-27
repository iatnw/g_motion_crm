-- name: GetDashboardTotals :one
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
      AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
      AND (sqlc.narg('exclude_cash')::boolean IS NOT TRUE OR d.is_cash_deal IS NOT TRUE)
),
deal_referral_totals AS (
    SELECT deal_id, SUM(amount)::numeric AS referral_total
    FROM referrals
    WHERE deal_id IN (SELECT id FROM filtered_deals)
    GROUP BY deal_id
)
SELECT
    COUNT(*)::bigint AS total_deals,
    COALESCE(SUM(fd.broker_fee), 0::numeric)::numeric AS gross_revenue,
    COALESCE(SUM(fd.broker_commission), 0::numeric)::numeric AS broker_payout,
    COALESCE(SUM(fd.doc_fee), 0::numeric)::numeric AS total_doc_fees,
    COALESCE(SUM(fd.your_net), 0::numeric)::numeric AS company_profit_before_tax,
    COALESCE(SUM(fd.broker_fee) FILTER (WHERE fd.deal_status = 'paid'), 0::numeric)::numeric AS paid_revenue,
    COALESCE(SUM(fd.broker_fee) FILTER (WHERE fd.deal_status = 'pending'), 0::numeric)::numeric AS pending_revenue,
    COALESCE(SUM(fd.broker_commission) FILTER (WHERE fd.broker_commission_status = 'paid'), 0::numeric)::numeric AS paid_broker_payout,
    COALESCE(SUM(fd.broker_commission) FILTER (WHERE fd.broker_commission_status = 'pending'), 0::numeric)::numeric AS unpaid_broker_payout,
    COALESCE(
        SUM(fd.your_net) FILTER (WHERE fd.deal_status = 'pending'),
        0::numeric
    )::numeric AS pending_company_profit,
    CASE
        WHEN COUNT(*) > 0 THEN COALESCE(SUM(fd.broker_fee), 0::numeric) / COUNT(*)
        ELSE 0::numeric
    END AS avg_revenue_per_deal,
    CASE
        WHEN COUNT(*) > 0
        THEN COALESCE(SUM(fd.your_net), 0::numeric) / COUNT(*)
        ELSE 0::numeric
    END AS avg_profit_per_deal,
    COALESCE(SUM(drt.referral_total), 0::numeric)::numeric AS total_referral_payout
FROM filtered_deals fd
LEFT JOIN deal_referral_totals drt ON drt.deal_id = fd.id;

-- name: GetDashboardAging :one
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
      AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
      AND d.deal_status = 'pending'
      AND d.deal_status_pending_since IS NOT NULL
      AND (sqlc.narg('exclude_cash')::boolean IS NOT TRUE OR d.is_cash_deal IS NOT TRUE)
      AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
),
aged AS (
    SELECT
        *,
        GREATEST((CURRENT_DATE - deal_status_pending_since), 0)::int AS pending_days
    FROM filtered_deals
)
SELECT
    COUNT(*) FILTER (WHERE pending_days <= 31)::bigint AS pending_0_31_count,
    COALESCE(SUM(broker_fee) FILTER (WHERE pending_days <= 31), 0::numeric)::numeric AS pending_0_31_revenue,
    COUNT(*) FILTER (WHERE pending_days BETWEEN 32 AND 40)::bigint AS pending_32_40_count,
    COALESCE(SUM(broker_fee) FILTER (WHERE pending_days BETWEEN 32 AND 40), 0::numeric)::numeric AS pending_32_40_revenue,
    COUNT(*) FILTER (WHERE pending_days > 40)::bigint AS pending_41_plus_count,
    COALESCE(SUM(broker_fee) FILTER (WHERE pending_days > 40), 0::numeric)::numeric AS pending_41_plus_revenue,
    COALESCE(MAX(pending_days), 0)::int AS oldest_pending_days
FROM aged;

-- name: ListDashboardBrokerStats :many
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
      AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
      AND d.broker_id IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
),
deal_referral_totals AS (
    SELECT deal_id, SUM(amount)::numeric AS referral_total
    FROM referrals
    WHERE deal_id IN (SELECT id FROM filtered_deals)
    GROUP BY deal_id
)
SELECT
    b.id,
    b.first_name,
    b.last_name,
    COUNT(fd.id)::bigint AS deal_count,
    COALESCE(SUM(fd.broker_fee), 0::numeric)::numeric AS gross_revenue,
    COALESCE(SUM(fd.broker_commission), 0::numeric)::numeric AS broker_payout,
    COALESCE(SUM(fd.your_net), 0::numeric)::numeric AS company_profit_before_tax,
    CASE
        WHEN COUNT(fd.id) > 0 THEN COALESCE(SUM(fd.broker_fee), 0::numeric) / COUNT(fd.id)
        ELSE 0::numeric
    END AS avg_revenue_per_deal
FROM brokers b
JOIN filtered_deals fd ON fd.broker_id = b.id
LEFT JOIN deal_referral_totals drt ON drt.deal_id = fd.id
GROUP BY b.id, b.first_name, b.last_name
ORDER BY gross_revenue DESC, deal_count DESC;

-- name: ListDashboardDealerStats :many
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
      AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
      AND d.dealer_id IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
)
SELECT
    d.id,
    d.name,
    COUNT(fd.id)::bigint AS deal_count,
    COALESCE(SUM(fd.broker_fee), 0::numeric)::numeric AS gross_revenue,
    AVG(
        CASE
            WHEN fd.broker_fee_paid_date IS NOT NULL
             AND fd.deal_date IS NOT NULL
            THEN (fd.broker_fee_paid_date - fd.deal_date)::numeric
        END
    )::numeric AS avg_days_to_payment
FROM dealers d
JOIN filtered_deals fd ON fd.dealer_id = d.id
GROUP BY d.id, d.name
ORDER BY gross_revenue DESC, deal_count DESC;

-- name: ListDashboardRecentDeals :many
SELECT d.*
FROM deals d
WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
  AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
  AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
  AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
ORDER BY COALESCE(d.deal_date, d.created_at::date) DESC, d.created_at DESC
LIMIT $1;

-- name: ListDashboardLeases :many
SELECT *
FROM (
    SELECT
        d.id,
        d.client_id,
        d.deal_number,
        d.deal_status,
        d.make,
        d.model,
        d.deal_date,
        (c.first_name || ' ' || c.last_name)::text AS client_name,
        d.lease_term_months,
        (d.lease_term_months - (EXTRACT(YEAR FROM age(CURRENT_DATE, COALESCE(d.deal_date, d.created_at::date))) * 12 + EXTRACT(MONTH FROM age(CURRENT_DATE, COALESCE(d.deal_date, d.created_at::date))))::int)::int AS lease_months_remaining
    FROM deals d
    JOIN clients c ON c.id = d.client_id
    WHERE COALESCE(d.deal_status != 'cancelled', TRUE)
      AND NULLIF(BTRIM(d.vin), '') IS NOT NULL
      AND d.lease_term_months IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR COALESCE(d.deal_date, d.created_at::date) <= sqlc.narg('to_date')::date)
      AND (sqlc.narg('exclude_cash')::boolean IS NOT TRUE OR d.is_cash_deal IS NOT TRUE)
) leases
ORDER BY (lease_months_remaining <= 0) ASC, lease_months_remaining ASC, deal_date DESC
LIMIT 20;
