-- name: CreateClient :one
INSERT INTO clients(
    first_name,
    last_name,
    phone,
    email,
    source,
    status,
    notes,
    date_of_birth,
    address,
    city,
    state,
    zip,
    ssn,
    housing_status,
    housing_duration,
    housing_monthly_payment,
    employer_name,
    employer_address,
    employer_city,
    employer_state,
    employer_zip,
    occupation,
    employer_phone,
    employment_status,
    employment_duration,
    gross_monthly_income,
    other_monthly_income,
    other_income_source
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
) RETURNING *;

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
    first_name              = $2,
    last_name               = $3,
    phone                   = $4,
    email                   = $5,
    source                  = $6,
    status                  = $7,
    notes                   = $8,
    date_of_birth           = $9,
    address                 = $10,
    city                    = $11,
    state                   = $12,
    zip                     = $13,
    ssn                     = $14,
    housing_status          = $15,
    housing_duration        = $16,
    housing_monthly_payment = $17,
    employer_name           = $18,
    employer_address        = $19,
    employer_city           = $20,
    employer_state          = $21,
    employer_zip            = $22,
    occupation              = $23,
    employer_phone          = $24,
    employment_status       = $25,
    employment_duration     = $26,
    gross_monthly_income    = $27,
    other_monthly_income    = $28,
    other_income_source     = $29,
    updated_at              = NOW()
WHERE id = $1
RETURNING *;
