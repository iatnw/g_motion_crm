-- +goose Up

WITH numbered AS (
    SELECT
        id,
        ROW_NUMBER() OVER (ORDER BY created_at, id) AS seq
    FROM deals
)
UPDATE deals d
SET deal_number = 'GM-' || LPAD(numbered.seq::text, 4, '0')
FROM numbered
WHERE d.id = numbered.id;

SELECT setval(
    'deal_number_seq',
    GREATEST((SELECT COUNT(*) FROM deals), 1),
    (SELECT COUNT(*) FROM deals) > 0
);

CREATE UNIQUE INDEX IF NOT EXISTS deals_deal_number_unique
ON deals(deal_number)
WHERE deal_number IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS deals_deal_number_unique;
