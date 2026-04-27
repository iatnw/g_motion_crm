-- +goose Up

CREATE TYPE deal_status AS ENUM ('paid', 'pending', 'cancelled');
CREATE TYPE broker_payment_method AS ENUM ('zelle_gmotion', 'zelle_personal');

ALTER TABLE deals
ADD COLUMN deal_status deal_status,
ADD COLUMN collect_amount NUMERIC,
ADD COLUMN doc_fee NUMERIC,
ADD COLUMN broker_payment_method broker_payment_method,
ADD COLUMN notes TEXT,
DROP COLUMN referral_amount,
DROP COLUMN referral_status,
DROP COLUMN referral_paid_date;

CREATE TABLE referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT,
    amount NUMERIC,
    status referral_status NOT NULL DEFAULT 'none',
    paid_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE referrals;

ALTER TABLE deals
    DROP COLUMN deal_status,
    DROP COLUMN collect_amount,
    DROP COLUMN doc_fee,
    DROP COLUMN broker_payment_method,
    DROP COLUMN notes,
    ADD COLUMN referral_amount NUMERIC,
    ADD COLUMN referral_status referral_status,
    ADD COLUMN referral_paid_date DATE;

DROP TYPE broker_payment_method;
DROP TYPE deal_status;
    
    

