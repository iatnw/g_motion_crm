-- +goose Up
-- Fix imported deals: pending_since should use deal_date, not the import date
UPDATE deals
SET deal_status_pending_since = deal_date
WHERE deal_status = 'pending'
  AND deal_date IS NOT NULL
  AND deal_status_pending_since IS NOT NULL
  AND deal_status_pending_since > deal_date;

-- +goose Down
