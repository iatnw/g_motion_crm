-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_deal_your_net(p_deal_id UUID)
RETURNS NUMERIC AS $$
    SELECT
        COALESCE(d.broker_fee, 0)
        + COALESCE(d.collect_amount, 0)
        - COALESCE(d.doc_fee, 0)
        - COALESCE(d.broker_commission, 0)
        - COALESCE(referral_totals.referral_total, 0)
    FROM deals d
    LEFT JOIN (
        SELECT deal_id, SUM(amount)::numeric AS referral_total
        FROM referrals
        WHERE deal_id = p_deal_id
        GROUP BY deal_id
    ) referral_totals ON referral_totals.deal_id = d.id
    WHERE d.id = p_deal_id;
$$ LANGUAGE SQL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_deal_your_net()
RETURNS TRIGGER AS $$
DECLARE
    referral_total NUMERIC;
BEGIN
    SELECT COALESCE(SUM(amount), 0)
    INTO referral_total
    FROM referrals
    WHERE deal_id = NEW.id;

    NEW.your_net =
        COALESCE(NEW.broker_fee, 0)
        + COALESCE(NEW.collect_amount, 0)
        - COALESCE(NEW.doc_fee, 0)
        - COALESCE(NEW.broker_commission, 0)
        - referral_total;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION refresh_deal_your_net_from_referral()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP IN ('UPDATE', 'DELETE') AND OLD.deal_id IS NOT NULL THEN
        UPDATE deals
        SET your_net = calculate_deal_your_net(OLD.deal_id),
            updated_at = NOW()
        WHERE id = OLD.deal_id;
    END IF;

    IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.deal_id IS NOT NULL THEN
        UPDATE deals
        SET your_net = calculate_deal_your_net(NEW.deal_id),
            updated_at = NOW()
        WHERE id = NEW.deal_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

UPDATE deals
SET your_net = calculate_deal_your_net(id);

CREATE TRIGGER trg_set_deal_your_net
BEFORE INSERT OR UPDATE OF broker_fee, collect_amount, doc_fee, broker_commission
ON deals
FOR EACH ROW
EXECUTE FUNCTION set_deal_your_net();

CREATE TRIGGER trg_refresh_deal_your_net_from_referral
AFTER INSERT OR UPDATE OF deal_id, amount OR DELETE
ON referrals
FOR EACH ROW
EXECUTE FUNCTION refresh_deal_your_net_from_referral();

-- +goose Down

DROP TRIGGER IF EXISTS trg_refresh_deal_your_net_from_referral ON referrals;
DROP TRIGGER IF EXISTS trg_set_deal_your_net ON deals;
DROP FUNCTION IF EXISTS refresh_deal_your_net_from_referral();
DROP FUNCTION IF EXISTS set_deal_your_net();
DROP FUNCTION IF EXISTS calculate_deal_your_net(UUID);
