-- +goose Up

ALTER TABLE tally_webhook_events ADD COLUMN processed_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE tally_webhook_events DROP COLUMN processed_at;
