-- +goose Up
ALTER TYPE payment_status RENAME VALUE 'not_paid' TO 'pending';
ALTER TYPE referral_status RENAME VALUE 'not_paid' TO 'pending';

-- +goose Down
ALTER TYPE payment_status RENAME VALUE 'pending' TO 'not_paid';
ALTER TYPE referral_status RENAME VALUE 'pending' TO 'not_paid';