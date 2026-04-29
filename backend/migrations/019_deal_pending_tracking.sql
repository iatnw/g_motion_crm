-- +goose Up

ALTER TABLE deals
ADD COLUMN deal_status_pending_since DATE,
ADD COLUMN payment_reminder_email_sent_at TIMESTAMPTZ;

UPDATE deals
SET deal_status_pending_since = CURRENT_DATE
WHERE deal_status = 'pending'
  AND deal_status_pending_since IS NULL;

-- +goose Down

ALTER TABLE deals
DROP COLUMN payment_reminder_email_sent_at,
DROP COLUMN deal_status_pending_since;
