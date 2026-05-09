-- +goose Up

ALTER TABLE tally_webhook_events ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ;

-- +goose Down
-- intentional no-op: processed_at predates this migration (added during a hot-reload cycle
-- before the migration file existed). Dropping it here would break the Tally service code
-- that reads and writes processed_at.
