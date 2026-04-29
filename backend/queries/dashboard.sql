-- name: GetDashboardTotals :one
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    JOIN clients c ON c.id = d.client_id
    WHERE c.status != 'no_deal'
      AND COALESCE(d.deal_status != 'cancelled', TRUE)
      AND (sqlc.narg('from_date')::date IS NULL OR d.deal_date >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR d.deal_date <= sqlc.narg('to_date')::date)
)
SELECT
    COUNT(*)::bigint AS total_deals,
    COALESCE(SUM(broker_fee), 0::numeric)::numeric AS gross_revenue,
    COALESCE(SUM(broker_commission), 0::numeric)::numeric AS broker_payout,
    COALESCE(SUM(doc_fee), 0::numeric)::numeric AS total_doc_fees,
    COALESCE(SUM(COALESCE(broker_fee, 0) - COALESCE(broker_commission, 0) - COALESCE(doc_fee, 0)), 0::numeric)::numeric AS company_profit_before_tax,
    COALESCE(SUM(broker_fee) FILTER (WHERE broker_fee_status = 'paid'), 0::numeric)::numeric AS paid_revenue,
    COALESCE(SUM(broker_fee) FILTER (WHERE broker_fee_status = 'pending'), 0::numeric)::numeric AS pending_revenue,
    COALESCE(SUM(broker_commission) FILTER (WHERE broker_commission_status = 'paid'), 0::numeric)::numeric AS paid_broker_payout,
    COALESCE(SUM(broker_commission) FILTER (WHERE broker_commission_status = 'pending'), 0::numeric)::numeric AS unpaid_broker_payout,
    COALESCE(
        SUM(COALESCE(broker_fee, 0) - COALESCE(broker_commission, 0) - COALESCE(doc_fee, 0))
            FILTER (WHERE broker_fee_status = 'pending'),
        0::numeric
    )::numeric AS pending_company_profit,
    CASE
        WHEN COUNT(*) > 0 THEN COALESCE(SUM(broker_fee), 0::numeric) / COUNT(*)
        ELSE 0::numeric
    END AS avg_revenue_per_deal,
    CASE
        WHEN COUNT(*) > 0
        THEN COALESCE(SUM(COALESCE(broker_fee, 0) - COALESCE(broker_commission, 0) - COALESCE(doc_fee, 0)), 0::numeric) / COUNT(*)
        ELSE 0::numeric
    END AS avg_profit_per_deal
FROM filtered_deals;

-- name: GetDashboardAging :one
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    JOIN clients c ON c.id = d.client_id
    WHERE c.status != 'no_deal'
      AND COALESCE(d.deal_status != 'cancelled', TRUE)
      AND d.deal_status = 'pending'
      AND d.deal_status_pending_since IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR d.deal_date >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR d.deal_date <= sqlc.narg('to_date')::date)
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
    JOIN clients c ON c.id = d.client_id
    WHERE c.status != 'no_deal'
      AND COALESCE(d.deal_status != 'cancelled', TRUE)
      AND d.broker_id IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR d.deal_date >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR d.deal_date <= sqlc.narg('to_date')::date)
)
SELECT
    b.id,
    b.first_name,
    b.last_name,
    COUNT(fd.id)::bigint AS deal_count,
    COALESCE(SUM(fd.broker_fee), 0::numeric)::numeric AS gross_revenue,
    COALESCE(SUM(fd.broker_commission), 0::numeric)::numeric AS broker_payout,
    COALESCE(SUM(COALESCE(fd.broker_fee, 0) - COALESCE(fd.broker_commission, 0) - COALESCE(fd.doc_fee, 0)), 0::numeric)::numeric AS company_profit_before_tax,
    CASE
        WHEN COUNT(fd.id) > 0 THEN COALESCE(SUM(fd.broker_fee), 0::numeric) / COUNT(fd.id)
        ELSE 0::numeric
    END AS avg_revenue_per_deal
FROM brokers b
JOIN filtered_deals fd ON fd.broker_id = b.id
GROUP BY b.id, b.first_name, b.last_name
ORDER BY gross_revenue DESC, deal_count DESC;

-- name: ListDashboardDealerStats :many
WITH filtered_deals AS (
    SELECT d.*
    FROM deals d
    JOIN clients c ON c.id = d.client_id
    WHERE c.status != 'no_deal'
      AND COALESCE(d.deal_status != 'cancelled', TRUE)
      AND d.dealer_id IS NOT NULL
      AND (sqlc.narg('from_date')::date IS NULL OR d.deal_date >= sqlc.narg('from_date')::date)
      AND (sqlc.narg('to_date')::date IS NULL OR d.deal_date <= sqlc.narg('to_date')::date)
)
SELECT
    d.id,
    d.name,
    COUNT(fd.id)::bigint AS deal_count,
    COALESCE(SUM(fd.broker_fee), 0::numeric)::numeric AS gross_revenue,
    COALESCE(AVG(
        CASE
            WHEN fd.broker_fee_paid_date IS NOT NULL
             AND fd.deal_date IS NOT NULL
            THEN (fd.broker_fee_paid_date - fd.deal_date)::numeric
        END
    ), 0::numeric)::numeric AS avg_days_to_payment
FROM dealers d
JOIN filtered_deals fd ON fd.dealer_id = d.id
GROUP BY d.id, d.name
ORDER BY gross_revenue DESC, deal_count DESC;

-- name: ListDashboardRecentDeals :many
SELECT d.*
FROM deals d
JOIN clients c ON c.id = d.client_id
WHERE c.status != 'no_deal'
  AND COALESCE(d.deal_status != 'cancelled', TRUE)
  AND (sqlc.narg('from_date')::date IS NULL OR d.deal_date >= sqlc.narg('from_date')::date)
  AND (sqlc.narg('to_date')::date IS NULL OR d.deal_date <= sqlc.narg('to_date')::date)
ORDER BY d.deal_date DESC NULLS LAST, d.created_at DESC
LIMIT $1;
