-- name: CreateSetting :one
INSERT INTO settings (
    owner_password_hash,
    owner_totp_secret
) VALUES ($1, $2) RETURNING *;

-- name: GetSetting :one
SELECT * FROM settings
WHERE id = 1;

-- name: UpdateOwnerPassword :one
UPDATE settings
SET
    owner_password_hash = $1,
    updated_at = NOW()
WHERE id = 1
RETURNING *;

-- name: SetupOwnerPassword :execresult
UPDATE settings
SET
    owner_password_hash = $1,
    updated_at = NOW()
WHERE id = 1 AND owner_password_hash = '';

-- name: UpdateOwnerTOTP :one
UPDATE settings
SET
    owner_totp_secret = $1,
    updated_at = NOW()
WHERE id = 1
RETURNING *;
