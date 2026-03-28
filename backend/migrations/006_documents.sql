-- +goose Up

CREATE TYPE document_type AS ENUM(
    'credit_app_pdf', 
    'driver_license',
    'insurance', 
    'broker_agreement',
    'dealer_docs',
    'check_copy', 
    'other'
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID REFERENCES deals(id),
    client_id UUID REFERENCES clients(id),
    s3_key TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size_bytes INTEGER,
    document_type document_type NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT must_have_owner CHECK (deal_id IS NOT NULL OR client_id IS NOT NULL)
);

-- +goose Down

DROP TABLE documents;
DROP TYPE document_type;
