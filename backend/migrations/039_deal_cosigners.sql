-- +goose Up

CREATE TABLE deal_cosigners (
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (deal_id, client_id)
);

CREATE INDEX idx_deal_cosigners_client_id ON deal_cosigners(client_id);

-- +goose Down

DROP TABLE deal_cosigners;
