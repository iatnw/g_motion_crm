-- +goose Up

CREATE TABLE tally_document_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tally_event_id BIGINT NOT NULL REFERENCES tally_webhook_events(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id),
    deal_id UUID REFERENCES deals(id),
    source_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    document_type document_type NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    document_id UUID REFERENCES documents(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE (tally_event_id, source_url, file_name, document_type)
);

CREATE INDEX idx_tally_document_imports_pending
    ON tally_document_imports (status, updated_at)
    WHERE status != 'completed';

-- +goose Down

DROP TABLE tally_document_imports;
