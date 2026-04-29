-- +goose Up

CREATE OR REPLACE VIEW dealer_stats AS
SELECT
    d.id,
    COUNT(dl.id) AS deal_count,
    COALESCE(SUM(dl.broker_fee), 0::numeric) AS total_amount_to_company,
    AVG(
        CASE
            WHEN dl.broker_fee_paid_date IS NOT NULL
             AND dl.deal_date IS NOT NULL
            THEN (dl.broker_fee_paid_date - dl.deal_date)::numeric
        END
    ) AS avg_days_to_payment
FROM dealers d
LEFT JOIN (
    SELECT deals.*
    FROM deals
    JOIN clients ON clients.id = deals.client_id
    WHERE clients.status != 'no_deal'
) dl ON dl.dealer_id = d.id
GROUP BY d.id;

-- +goose Down

CREATE OR REPLACE VIEW dealer_stats AS
SELECT
    d.id,
    COUNT(dl.id) AS deal_count,
    COALESCE(SUM(dl.broker_fee), 0::numeric) AS total_broker_fee,
    AVG(
        CASE
            WHEN dl.broker_fee_paid_date IS NOT NULL
             AND dl.deal_date IS NOT NULL
            THEN (dl.broker_fee_paid_date - dl.deal_date)::numeric
        END
    ) AS avg_days_to_payment
FROM dealers d
LEFT JOIN deals dl ON dl.dealer_id = d.id
GROUP BY d.id;
