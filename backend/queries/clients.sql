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

-- name: GetClientByEmail :one
SELECT *
FROM clients
WHERE email = $1::text
LIMIT 1;;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1;

-- name: ListClients :many
SELECT * FROM clients
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

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
    ssn                     = COALESCE($14, ssn),
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

-- name: MergeClientFromTally :one
UPDATE clients
SET
    first_name              = $2,
    last_name               = $3,
    phone                   = COALESCE($4, phone),
    email                   = COALESCE($5, email),
    date_of_birth           = COALESCE($6, date_of_birth),
    address                 = COALESCE($7, address),
    city                    = COALESCE($8, city),
    state                   = COALESCE($9, state),
    zip                     = COALESCE($10, zip),
    ssn                     = COALESCE($11, ssn),
    housing_status          = COALESCE($12, housing_status),
    housing_duration        = COALESCE($13, housing_duration),
    housing_monthly_payment = COALESCE($14, housing_monthly_payment),
    employer_name           = COALESCE($15, employer_name),
    employer_address        = COALESCE($16, employer_address),
    employer_city           = COALESCE($17, employer_city),
    employer_state          = COALESCE($18, employer_state),
    employer_zip            = COALESCE($19, employer_zip),
    occupation              = COALESCE($20, occupation),
    employer_phone          = COALESCE($21, employer_phone),
    employment_status       = COALESCE($22, employment_status),
    employment_duration     = COALESCE($23, employment_duration),
    gross_monthly_income    = COALESCE($24, gross_monthly_income),
    other_monthly_income    = COALESCE($25, other_monthly_income),
    other_income_source     = COALESCE($26, other_income_source),
    updated_at              = NOW()
WHERE id = $1
RETURNING *;
