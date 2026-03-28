-- name: CreateDeal :one
INSERT INTO deals (
    client_id,
    broker_id,
    dealer_id,
    status,
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
    referral_amount,
    your_net,
    broker_fee_status,
    broker_fee_paid_date,
    referral_status,
    referral_paid_date
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
$17, $18, $19, $20, $21, $22, $23, $24) RETURNING *;

-- name: GetDeal :one
SELECT * FROM deals
WHERE id = $1;

-- name: ListDeals :many
SELECT * FROM deals
ORDER BY created_at DESC;

-- name: UpdateDeal :one
UPDATE deals
SET 
    client_id = $2,
    broker_id = $3,
    dealer_id = $4,
    status = $5,
    deal_date = $6,
    deal_number = $7,
    vin = $8,
    make = $9,
    model = $10,
    trim_level = $11,
    down_payment = $12,
    down_payment_method = $13,
    monthly_payment = $14,
    broker_fee = $15,
    broker_percent = $16,
    broker_commission = $17,
    broker_commission_status = $18,
    broker_commission_paid_date = $19,
    referral_amount = $20,
    your_net = $21,
    broker_fee_status = $22,
    broker_fee_paid_date = $23,
    referral_status = $24,
    referral_paid_date = $25,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListDealsByClient :many
SELECT * FROM deals
WHERE deals.client_id = $1
ORDER BY created_at DESC;

-- name: UpdateDealBrokerFeeStatus :one
UPDATE deals
SET
    broker_fee_status = $2,
    broker_fee_paid_date = CASE WHEN $2::payment_status = 'paid' 
THEN CURRENT_DATE ELSE NULL END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDealBrokerCommissionStatus :one
UPDATE deals
SET
    broker_commission_status = $2,
    broker_commission_paid_date = CASE WHEN $2::payment_status = 'paid' 
THEN CURRENT_DATE ELSE NULL END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDealReferralStatus :one
UPDATE deals
SET
    referral_status = $2,
    referral_paid_date = CASE WHEN $2::referral_status = 'paid'
THEN CURRENT_DATE ELSE NULL END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


