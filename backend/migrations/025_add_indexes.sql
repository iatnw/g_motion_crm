-- +goose Up

CREATE INDEX idx_deals_client_id   ON deals(client_id);
CREATE INDEX idx_deals_broker_id   ON deals(broker_id);
CREATE INDEX idx_deals_dealer_id   ON deals(dealer_id);
CREATE INDEX idx_deals_deal_status ON deals(deal_status);
CREATE INDEX idx_deals_created_at  ON deals(created_at DESC);

CREATE INDEX idx_referrals_deal_id ON referrals(deal_id);

CREATE INDEX idx_documents_client_id ON documents(client_id);
CREATE INDEX idx_documents_deal_id   ON documents(deal_id);

-- +goose Down

DROP INDEX idx_deals_client_id;
DROP INDEX idx_deals_broker_id;
DROP INDEX idx_deals_dealer_id;
DROP INDEX idx_deals_deal_status;
DROP INDEX idx_deals_created_at;

DROP INDEX idx_referrals_deal_id;

DROP INDEX idx_documents_client_id;
DROP INDEX idx_documents_deal_id;
