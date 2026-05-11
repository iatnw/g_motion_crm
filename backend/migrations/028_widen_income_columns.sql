-- +goose Up
ALTER TABLE clients
    ALTER COLUMN housing_monthly_payment TYPE NUMERIC(12,2),
    ALTER COLUMN gross_monthly_income    TYPE NUMERIC(12,2),
    ALTER COLUMN other_monthly_income    TYPE NUMERIC(12,2);

-- +goose Down
ALTER TABLE clients
    ALTER COLUMN housing_monthly_payment TYPE NUMERIC(8,2),
    ALTER COLUMN gross_monthly_income    TYPE NUMERIC(8,2),
    ALTER COLUMN other_monthly_income    TYPE NUMERIC(8,2);
