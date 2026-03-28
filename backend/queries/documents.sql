-- name: CreateDocument :one
INSERT INTO documents (
    deal_id,
    client_id,
    s3_key,
    file_name,
    file_size_bytes,
    document_type
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents
WHERE id = $1;

-- name: ListDocumentsByDeal :many
SELECT * FROM documents
WHERE documents.deal_id = $1;

-- name: ListDocumentsByClient :many
SELECT * FROM documents
WHERE documents.client_id = $1;

-- name: DeleteDocument :one
DELETE FROM documents
WHERE id = $1 
RETURNING *;
