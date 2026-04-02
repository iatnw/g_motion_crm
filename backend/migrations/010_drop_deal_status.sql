-- +goose Up

ALTER TABLE deals DROP COLUMN status;
DROP TYPE deal_status;

-- +goose Down

CREATE TYPE deal_status AS ENUM ('draft', 'active', 'funded', 'cancelled');
ALTER TABLE deals ADD COLUMN status deal_status NOT NULL DEFAULT 'draft';
