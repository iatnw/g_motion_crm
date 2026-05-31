-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    entity_type,
    entity_id,
    actor,
    action, 
    changed_fields,
    ip_address,
    user_agent
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListAuditLogsByEntity :many
SELECT * FROM audit_logs
WHERE audit_logs.entity_id = $1 AND audit_logs.entity_type = $2
ORDER BY created_at DESC;
