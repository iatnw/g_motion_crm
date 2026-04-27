-- name: CreateDealer :one
INSERT INTO dealers (
    name,
    contact_person,
    phone,
    address,
    notes,
    contact_person_email
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetDealer :one
SELECT * FROM dealers
WHERE id = $1;

-- name: ListDealers :many
SELECT * FROM dealers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateDealer :one
UPDATE dealers
SET
    name = $2,
    contact_person = $3,
    phone = $4,
    address = $5,
    notes = $6,
    contact_person_email = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDealer :one
DELETE FROM dealers
WHERE id = $1
RETURNING *;

-- name: GetDealerStats :one
SELECT * FROM dealer_stats
WHERE id = $1;

-- name: ListDealerStats :many
SELECT * FROM dealer_stats;
