-- +goose Up
-- Repair deals whose pending clock was reset by an update/status change
-- instead of anchoring to deal_date (matches CreateDeal behavior).
UPDATE deals
SET deal_status_pending_since = deal_date
WHERE deal_status = 'pending'
  AND deal_date IS NOT NULL
  AND deal_status_pending_since IS NOT NULL
  AND deal_status_pending_since > deal_date;

-- +goose Down
-- Data repair cannot be undone.
SELECT 1;
