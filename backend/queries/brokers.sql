-- name: CreateBroker :one
INSERT INTO brokers (
    first_name,
    last_name,
    phone,
    email,
    notes
) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetBroker :one
SELECT * FROM brokers
WHERE id = $1;

-- name: ListBrokers :many
SELECT * FROM brokers
ORDER BY created_at DESC;

-- name: UpdateBroker :one
UPDATE brokers
SET
    first_name = $2,
    last_name = $3,
    phone = $4,
    email = $5,
    notes = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
