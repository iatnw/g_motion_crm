-- +goose Up

ALTER TABLE audit_logs ALTER COLUMN entity_id DROP NOT NULL;

CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id, created_at DESC);

-- +goose Down

DROP INDEX audit_logs_entity_idx;

DELETE FROM audit_logs WHERE entity_id IS NULL;

ALTER TABLE audit_logs ALTER COLUMN entity_id SET NOT NULL;
