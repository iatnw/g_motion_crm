-- +goose Up
ALTER TABLE clients ALTER COLUMN phone DROP NOT NULL;
ALTER TABLE clients ALTER COLUMN email DROP NOT NULL;

-- +goose Down
UPDATE clients SET phone = '' WHERE phone IS NULL;
UPDATE clients SET email = '' WHERE email IS NULL;
ALTER TABLE clients ALTER COLUMN phone SET NOT NULL;
ALTER TABLE clients ALTER COLUMN email SET NOT NULL;
