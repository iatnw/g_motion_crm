-- +goose Up

UPDATE deals
SET
    deal_status = 'pending',
    deal_status_pending_since = COALESCE(deal_date, CURRENT_DATE)
WHERE deal_status IS NULL;

UPDATE deals
SET broker_fee_status = 'pending'
WHERE broker_fee_status IS NULL
  AND deal_status IN ('pending', 'paid');

UPDATE deals
SET broker_commission_status = 'pending'
WHERE broker_commission_status IS NULL
  AND deal_status IN ('pending', 'paid');

UPDATE clients c
SET
    status = 'active',
    updated_at = NOW()
WHERE c.status = 'no_deal'
  AND EXISTS (
      SELECT 1
      FROM deals d
      WHERE d.client_id = c.id
        AND d.deal_status IN ('pending', 'paid')
  );

-- +goose Down

-- Data repair migration. Leaving existing rows unchanged is safer than
-- guessing which records were previously NULL/no_deal.
