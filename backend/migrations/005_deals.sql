-- +goose Up
CREATE TYPE deal_status AS ENUM ('draft','active','funded','cancelled'); 
CREATE TYPE down_payment_method AS ENUM ('cash', 'card');
CREATE TYPE payment_status AS ENUM ('paid', 'not_paid');
CREATE TYPE referral_status AS ENUM ('none', 'paid', 'not_paid');

CREATE TABLE deals(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id),
    broker_id UUID REFERENCES brokers(id),
    dealer_id UUID REFERENCES dealers(id),
    status deal_status NOT NULL DEFAULT 'draft',
    deal_date DATE,
    deal_number TEXT,
    vin TEXT,
    make TEXT,
    model TEXT,
    trim TEXT,
    down_payment NUMERIC(8,2),
    down_payment_method down_payment_method,
    monthly_payment NUMERIC(8,2),
    broker_fee NUMERIC(8,2),
    broker_percent NUMERIC(5,2),
    broker_commission NUMERIC(8,2),
    broker_commission_status payment_status DEFAULT 'not_paid',
    broker_commission_paid_date DATE,
    referral_amount NUMERIC(8,2),
    your_net NUMERIC(8,2),
    broker_fee_status payment_status DEFAULT 'not_paid',
    broker_fee_paid_date DATE,
    referral_status referral_status DEFAULT 'none',
    referral_paid_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose Down

DROP TABLE deals;
DROP TYPE deal_status;
DROP TYPE down_payment_method;
DROP TYPE payment_status;
DROP TYPE referral_status;
