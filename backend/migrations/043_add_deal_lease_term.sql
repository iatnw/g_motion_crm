-- +goose Up

ALTER TABLE deals
ADD COLUMN lease_term_months INTEGER;

-- +goose Down

ALTER TABLE deals
DROP COLUMN lease_term_months;
