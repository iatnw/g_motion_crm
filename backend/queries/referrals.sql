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
-- $1 is nullable: pass pgtype.UUID{} (Valid:false) when no deal

-- name: GetReferral :one
SELECT * FROM referrals
WHERE id = $1;

-- name: ListReferralsByDeal :many
SELECT * FROM referrals
WHERE deal_id = $1
ORDER BY created_at DESC;

-- name: ListReferralsByDeals :many
SELECT * FROM referrals
WHERE deal_id = ANY($1::uuid[])
ORDER BY created_at DESC;

-- name: UpdateReferral :one
UPDATE referrals
SET
    deal_id    = $2,
    first_name = $3,
    last_name  = $4,
    phone      = $5,
    amount     = $6,
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

-- name: ListAllReferrals :many
SELECT
    r.id,
    r.deal_id,
    r.first_name,
    r.last_name,
    r.phone,
    r.amount,
    r.status,
    r.paid_date,
    r.notes,
    r.created_at,
    r.updated_at,
    d.deal_number
FROM referrals r
LEFT JOIN deals d ON d.id = r.deal_id
ORDER BY r.created_at DESC;
