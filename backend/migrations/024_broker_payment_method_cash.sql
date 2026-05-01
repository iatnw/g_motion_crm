-- +goose Up
ALTER TYPE broker_payment_method ADD VALUE IF NOT EXISTS 'cash';

-- +goose Down
-- Postgres does not support removing enum values; down is a no-op
