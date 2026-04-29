-- +goose Up

CREATE OR REPLACE VIEW broker_stats AS
SELECT
    b.id,
    COUNT(dl.id) AS total_deals,
    COUNT(dl.id) FILTER (
        WHERE DATE_TRUNC('month', dl.deal_date) = DATE_TRUNC('month', CURRENT_DATE)
        ) AS current_month_deals,
    COALESCE(SUM(dl.broker_commission), 0::numeric) AS total_broker_commission,
    COALESCE(SUM(dl.broker_fee), 0::numeric) AS total_broker_fee,
    (COALESCE(SUM(dl.broker_fee), 0::numeric) - COALESCE(SUM(dl.broker_commission), 0::numeric)) AS company_net_income,
    CASE
        WHEN COUNT(dl.id) > 0
        THEN SUM(dl.broker_fee) / COUNT(dl.id)
        ELSE 0::numeric
    END AS avg_broker_fee_per_deal,
    CASE
        WHEN COUNT(DISTINCT DATE_TRUNC('month', dl.deal_date)) > 0
        THEN COUNT(dl.id)::numeric / COUNT(DISTINCT DATE_TRUNC('month', dl.deal_date))
        ELSE 0::numeric
    END AS avg_deals_per_month
FROM brokers b
LEFT JOIN deals dl ON dl.broker_id = b.id
GROUP BY b.id;

-- +goose Down

CREATE OR REPLACE VIEW broker_stats AS
SELECT
    b.id,
    COUNT(dl.id) AS total_deals,
    COUNT(dl.id) FILTER (
        WHERE DATE_TRUNC('month', dl.deal_date) = DATE_TRUNC('month', CURRENT_DATE)
        ) AS current_month_deals,
    COALESCE(SUM(dl.broker_commission), 0::numeric) AS total_broker_commission,
    COALESCE(SUM(dl.broker_fee), 0::numeric) AS total_broker_fee,
    (COALESCE(SUM(dl.broker_fee), 0::numeric) - COALESCE(SUM(dl.broker_commission), 0::numeric)) AS broker_clean_income,
    CASE
        WHEN COUNT(dl.id) > 0
        THEN SUM(dl.broker_fee) / COUNT(dl.id)
        ELSE 0::numeric
    END AS avg_broker_fee_per_deal,
    CASE
        WHEN COUNT(DISTINCT DATE_TRUNC('month', dl.deal_date)) > 0
        THEN COUNT(dl.id)::numeric / COUNT(DISTINCT DATE_TRUNC('month', dl.deal_date))
        ELSE 0::numeric
    END AS avg_deals_per_month
FROM brokers b
LEFT JOIN deals dl ON dl.broker_id = b.id
GROUP BY b.id;
