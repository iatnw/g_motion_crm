-- name: CreateDealer :one
INSERT INTO dealers (
    name,
    contact_person,
    phone,
    address,
    notes
) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetDealer :one
SELECT * FROM dealers
WHERE id = $1;

-- name: ListDealers :many
SELECT * FROM dealers
ORDER BY created_at DESC;

-- name: UpdateDealer :one
UPDATE dealers
SET
    name = $2,
    contact_person = $3,
    phone = $4,
    address = $5,
    notes = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
