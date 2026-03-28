-- +goose Up

ALTER TABLE deals RENAME COLUMN trim TO trim_level;

-- +goose Down 

ALTER TABLE deals RENAME COLUMN trim_level TO trim;
