-- +goose Up

CREATE TYPE clients_source AS ENUM(
    'instagram',
    'telegram',
    'referall',
    'organic',
    'repeat_client'
);

CREATE TYPE client_status AS ENUM(
    'active',
    'closed',
    'no_deal'
);

CREATE TABLE clients(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    source clients_source,
    status client_status NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clients_phone ON clients(phone);

-- +goose Down

DROP TABLE clients;
DROP TYPE client_status;
DROP TYPE clients_source;
