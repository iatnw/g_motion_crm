-- +goose Up

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_deals_make_trgm ON deals USING gin (make gin_trgm_ops);
CREATE INDEX idx_deals_model_trgm ON deals USING gin (model gin_trgm_ops);
CREATE INDEX idx_deals_vin_trgm ON deals USING gin (vin gin_trgm_ops);
CREATE INDEX idx_deals_deal_number_trgm ON deals USING gin (deal_number gin_trgm_ops);

CREATE INDEX idx_clients_first_name_trgm ON clients USING gin (first_name gin_trgm_ops);
CREATE INDEX idx_clients_last_name_trgm ON clients USING gin (last_name gin_trgm_ops);
CREATE INDEX idx_clients_full_name_trgm ON clients USING gin (((first_name || ' ' || last_name)) gin_trgm_ops);

CREATE INDEX idx_brokers_first_name_trgm ON brokers USING gin (first_name gin_trgm_ops);
CREATE INDEX idx_brokers_last_name_trgm ON brokers USING gin (last_name gin_trgm_ops);
CREATE INDEX idx_brokers_full_name_trgm ON brokers USING gin (((first_name || ' ' || last_name)) gin_trgm_ops);

-- +goose Down

DROP INDEX IF EXISTS idx_brokers_full_name_trgm;
DROP INDEX IF EXISTS idx_brokers_last_name_trgm;
DROP INDEX IF EXISTS idx_brokers_first_name_trgm;

DROP INDEX IF EXISTS idx_clients_full_name_trgm;
DROP INDEX IF EXISTS idx_clients_last_name_trgm;
DROP INDEX IF EXISTS idx_clients_first_name_trgm;

DROP INDEX IF EXISTS idx_deals_deal_number_trgm;
DROP INDEX IF EXISTS idx_deals_vin_trgm;
DROP INDEX IF EXISTS idx_deals_model_trgm;
DROP INDEX IF EXISTS idx_deals_make_trgm;
