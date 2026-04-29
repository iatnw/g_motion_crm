-- +goose Up

CREATE TABLE tally_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT UNIQUE,
    submission_id TEXT UNIQUE,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT has_tally_identifier CHECK (
        event_id IS NOT NULL OR submission_id IS NOT NULL
    )
);

-- +goose Down

DROP TABLE tally_webhook_events;
