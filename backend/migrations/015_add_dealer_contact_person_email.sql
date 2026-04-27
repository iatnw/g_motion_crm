-- +goose Up

ALTER TABLE dealers ADD COLUMN contact_person_email TEXT;

-- +goose Down

ALTER TABLE dealers DROP COLUMN contact_person_email;
