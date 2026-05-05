-- +goose Up

-- The individual UNIQUE constraints on event_id and submission_id allow multiple
-- NULLs (PostgreSQL treats each NULL as distinct). Adding a composite index with
-- COALESCE ensures the combination (event_id, submission_id) is also unique when
-- either or both fields are NULL, closing the duplicate-processing gap.
CREATE UNIQUE INDEX idx_tally_webhook_events_dedup
    ON tally_webhook_events (COALESCE(event_id, ''), COALESCE(submission_id, ''));

-- +goose Down

DROP INDEX idx_tally_webhook_events_dedup;
