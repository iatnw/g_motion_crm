-- +goose Up

ALTER TABLE audit_logs ADD COLUMN actor TEXT NOT NULL DEFAULT 'owner';

-- +goose Down

ALTER TABLE audit_logs DROP COLUMN actor;
