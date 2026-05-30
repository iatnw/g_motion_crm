-- +goose Up

CREATE INDEX idx_dealers_name_trgm ON dealers USING gin (name gin_trgm_ops);

-- +goose Down

DROP INDEX IF EXISTS idx_dealers_name_trgm;
