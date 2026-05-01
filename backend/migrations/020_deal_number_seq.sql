-- +goose Up
CREATE SEQUENCE IF NOT EXISTS deal_number_seq START 1;

-- +goose Down
DROP SEQUENCE IF EXISTS deal_number_seq;
