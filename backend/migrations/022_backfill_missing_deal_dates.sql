-- +goose Up

UPDATE deals
SET deal_date = created_at::date
WHERE deal_date IS NULL
  AND COALESCE(deal_status != 'cancelled', TRUE);

-- +goose Down

-- Data repair migration. We do not clear deal_date on rollback because some
-- rows may have been edited after the migration ran.
