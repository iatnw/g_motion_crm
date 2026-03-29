-- name: CreateClient :one
INSERT INTO clients(
    first_name,
    last_name,
    phone,
    email,
    source,
    status
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetClientByPhone :one
SELECT * FROM clients
WHERE phone = $1
LIMIT 1;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1;

-- name: ListClients :many
SELECT * FROM clients
ORDER BY created_at DESC;

-- name: UpdateClient :one
UPDATE clients
SET
    first_name = $2,
    last_name = $3,
    phone = $4,
    email = $5,
    source = $6,
    status = $7,
    updated_at = NOW()
WHERE id = $1 
RETURNING *;

