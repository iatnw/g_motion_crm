-- +goose Up

CREATE TYPE employment_status AS ENUM ('self_employed', 'full_time', 'part_time');

ALTER TABLE clients
    ADD COLUMN date_of_birth          DATE,
    ADD COLUMN address                TEXT,
    ADD COLUMN city                   TEXT,
    ADD COLUMN state                  TEXT,
    ADD COLUMN zip                    TEXT,
    ADD COLUMN ssn                    TEXT,
    ADD COLUMN housing_status         TEXT,
    ADD COLUMN housing_duration       TEXT,
    ADD COLUMN housing_monthly_payment NUMERIC(8,2),
    ADD COLUMN employer_name          TEXT,
    ADD COLUMN employer_address       TEXT,
    ADD COLUMN employer_city          TEXT,
    ADD COLUMN employer_state         TEXT,
    ADD COLUMN employer_zip           TEXT,
    ADD COLUMN occupation             TEXT,
    ADD COLUMN employer_phone         TEXT,
    ADD COLUMN employment_status      employment_status,
    ADD COLUMN employment_duration    TEXT,
    ADD COLUMN gross_monthly_income   NUMERIC(8,2),
    ADD COLUMN other_monthly_income   NUMERIC(8,2),
    ADD COLUMN other_income_source    TEXT;

-- +goose Down

ALTER TABLE clients
    DROP COLUMN date_of_birth,
    DROP COLUMN address,
    DROP COLUMN city,
    DROP COLUMN state,
    DROP COLUMN zip,
    DROP COLUMN ssn,
    DROP COLUMN housing_status,
    DROP COLUMN housing_duration,
    DROP COLUMN housing_monthly_payment,
    DROP COLUMN employer_name,
    DROP COLUMN employer_address,
    DROP COLUMN employer_city,
    DROP COLUMN employer_state,
    DROP COLUMN employer_zip,
    DROP COLUMN occupation,
    DROP COLUMN employer_phone,
    DROP COLUMN employment_status,
    DROP COLUMN employment_duration,
    DROP COLUMN gross_monthly_income,
    DROP COLUMN other_monthly_income,
    DROP COLUMN other_income_source;

DROP TYPE employment_status;
