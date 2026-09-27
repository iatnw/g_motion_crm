-- +goose Up
-- Pending deals must anchor to deal_date in both directions:
-- the clock always tracks how long money has been outstanding on the deal.
UPDATE deals
SET deal_status_pending_since = deal_date
WHERE deal_status = 'pending'
  AND deal_date IS NOT NULL
  AND deal_status_pending_since IS DISTINCT FROM deal_date;

-- +goose Down
-- Data repair cannot be undone.
SELECT 1;
