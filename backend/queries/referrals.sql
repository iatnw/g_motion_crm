-- name: CreateReferral :one
INSERT INTO referrals (
    deal_id,
    first_name,
    last_name,
    phone,
    amount,
    status,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetReferral :one
SELECT * FROM referrals
WHERE id = $1;

-- name: ListReferralsByDeal :many
SELECT * FROM referrals
WHERE deal_id = $1
ORDER BY created_at DESC;

-- name: UpdateReferral :one
UPDATE referrals
SET 
    first_name = $2,
    last_name  = $3,
    phone      = $4,
    amount     = $5,
    status     = $6,
    notes      = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateReferralStatus :one
UPDATE referrals
SET 
    status = $2,
    paid_date = CASE WHEN $2::referral_status = 'paid'
                THEN CURRENT_DATE ELSE NULL END,
    updated_at = NOW()
    WHERE id = $1
    RETURNING *;

-- name: DeleteReferral :one
DELETE FROM referrals
WHERE id = $1
RETURNING *;
