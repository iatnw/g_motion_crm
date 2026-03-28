-- +goose Up

CREATE TABLE settings (
    id INTEGER PRIMARY KEY DEFAULT 1,
    owner_password_hash TEXT NOT NULL DEFAULT '',
    owner_totp_secret TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id=1)
);

INSERT INTO settings DEFAULT VALUES;

-- +goose Down

DROP TABLE settings;
