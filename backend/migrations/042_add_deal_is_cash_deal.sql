-- +goose Up

ALTER TABLE deals
ADD COLUMN is_cash_deal BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down

ALTER TABLE deals
DROP COLUMN is_cash_deal;
