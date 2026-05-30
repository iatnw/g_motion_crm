-- +goose Up

CREATE UNIQUE INDEX idx_documents_s3_key_unique ON documents(s3_key);

-- +goose Down

DROP INDEX idx_documents_s3_key_unique;
