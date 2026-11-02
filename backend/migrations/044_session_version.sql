-- +goose Up
ALTER TABLE settings
ADD COLUMN session_version BIGINT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE settings DROP COLUMN session_version;
